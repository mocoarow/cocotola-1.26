package question

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// pendingAudioFinder loads questions in audioGeneration.status == pending.
type pendingAudioFinder interface {
	FindPendingAudio(ctx context.Context, limit int) ([]domainquestion.Question, error)
}

// staleGeneratingAudioFinder loads questions stuck in generating beyond a
// staleness threshold so the reaper can transition them back to pending.
type staleGeneratingAudioFinder interface {
	FindStaleGenerating(ctx context.Context, staleBefore time.Time, limit int) ([]domainquestion.Question, error)
}

// AudioBatchCommand exposes the internal lifecycle operations driven by the
// cocotola-audio-generator batch service: list pending, claim, complete, fail,
// reclaim-stale.
//
// All transitions go through Question domain methods (which enforce state
// invariants) and are persisted via Save (which applies optimistic locking on
// the question's version). Concurrent batch instances racing for the same item
// surface as ErrAudioConcurrentModification on the loser; callers should treat
// it the same as ErrAudioNotPending — skip the item, retry next run.
type AudioBatchCommand struct {
	finder  questionFinder
	pending pendingAudioFinder
	stale   staleGeneratingAudioFinder
	saver   questionSaver
	config  UsecaseConfig
}

// NewAudioBatchCommand returns a new AudioBatchCommand.
func NewAudioBatchCommand(finder questionFinder, pending pendingAudioFinder, stale staleGeneratingAudioFinder, saver questionSaver, config UsecaseConfig) *AudioBatchCommand {
	return &AudioBatchCommand{finder: finder, pending: pending, stale: stale, saver: saver, config: config}
}

// ListPendingAudio returns up to input.Limit pending queue entries.
func (c *AudioBatchCommand) ListPendingAudio(ctx context.Context, input *questionservice.ListPendingAudioInput) (*questionservice.ListPendingAudioOutput, error) {
	questions, err := c.pending.FindPendingAudio(ctx, input.Limit)
	if err != nil {
		return nil, fmt.Errorf("find pending audio: %w", err)
	}
	items := make([]questionservice.PendingAudioItem, 0, len(questions))
	for i := range questions {
		q := &questions[i]
		ag := q.AudioGeneration()
		if ag == nil {
			continue
		}
		if q.QuestionType().Value() != domainquestion.TypeWordFill().Value() {
			logSkippedNonWordFill(ctx, q)
			continue
		}
		content, err := domainquestion.ParseWordFillContent(q.Content())
		if err != nil {
			return nil, fmt.Errorf("parse word_fill content for question %s: %w", q.ID(), err)
		}
		items = append(items, questionservice.PendingAudioItem{
			WorkbookID:  q.WorkbookID(),
			QuestionID:  q.ID(),
			SourceText:  content.Source.Text,
			SourceLang:  content.Source.Lang,
			TargetText:  domainquestion.FillWordFillBlanks(content.Target.Text),
			TargetLang:  content.Target.Lang,
			InputHash:   ag.InputHash(),
			FailedTries: ag.FailedAttempts(),
			UpdatedAt:   ag.UpdatedAt(),
		})
	}
	return &questionservice.ListPendingAudioOutput{Items: items}, nil
}

// ClaimAudio transitions a question's audio from pending to generating.
//
// Returns:
//   - domain.ErrQuestionNotFound when the question does not exist.
//   - domain.ErrAudioNotPending when the audio is not in pending state.
//   - domain.ErrAudioInputHashMismatch when the input hash does not match.
//   - domain.ErrAudioConcurrentModification when the optimistic-lock check
//     fails (another writer modified the question between load and save).
func (c *AudioBatchCommand) ClaimAudio(ctx context.Context, input *questionservice.ClaimAudioInput) error {
	q, err := c.finder.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return fmt.Errorf("find question: %w", err)
	}
	if err := q.ClaimAudio(input.InputHash, c.config.Now()); err != nil {
		return fmt.Errorf("claim audio: %w", err)
	}
	if err := c.saver.Save(ctx, q); err != nil {
		return mapAudioSaveError("claim audio save", err)
	}
	return nil
}

// CompleteAudio transitions a question's audio from generating to ready,
// persisting the generated refs.
func (c *AudioBatchCommand) CompleteAudio(ctx context.Context, input *questionservice.CompleteAudioInput) error {
	q, err := c.finder.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return fmt.Errorf("find question: %w", err)
	}
	refs := make(map[string]domainquestion.AudioRef, len(input.Refs))
	for k, v := range input.Refs {
		ref, err := domainquestion.NewAudioRef(v.Path, v.DurationSec, v.SizeBytes)
		if err != nil {
			return fmt.Errorf("new audio ref %q: %w", k, err)
		}
		refs[k] = ref
	}
	if err := q.CompleteAudio(input.InputHash, refs, c.config.Now()); err != nil {
		return fmt.Errorf("complete audio: %w", err)
	}
	if err := c.saver.Save(ctx, q); err != nil {
		return mapAudioSaveError("complete audio save", err)
	}
	return nil
}

// FailAudio transitions a question's audio from generating to failed,
// incrementing the failed-attempts counter so the next batch can apply
// backoff.
func (c *AudioBatchCommand) FailAudio(ctx context.Context, input *questionservice.FailAudioInput) error {
	q, err := c.finder.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return fmt.Errorf("find question: %w", err)
	}
	if err := q.FailAudio(input.InputHash, input.Reason, c.config.Now()); err != nil {
		return fmt.Errorf("fail audio: %w", err)
	}
	if err := c.saver.Save(ctx, q); err != nil {
		return mapAudioSaveError("fail audio save", err)
	}
	return nil
}

// ReclaimStaleAudio transitions questions stuck in generating beyond
// staleAfter back to pending so they get re-queued. Concurrent reapers and
// regular Claim calls are protected by optimistic locking — losers surface as
// ErrAudioConcurrentModification and are silently skipped (the work is
// idempotent).
//
// Returns the number of items successfully reclaimed.
func (c *AudioBatchCommand) ReclaimStaleAudio(ctx context.Context, input *questionservice.ReclaimStaleAudioInput) (*questionservice.ReclaimStaleAudioOutput, error) {
	now := c.config.Now()
	staleBefore := now.Add(-input.StaleAfter)
	questions, err := c.stale.FindStaleGenerating(ctx, staleBefore, input.Limit)
	if err != nil {
		return nil, fmt.Errorf("find stale generating: %w", err)
	}
	reclaimed := 0
	for i := range questions {
		q := &questions[i]
		if !q.ReclaimStaleAudio(now, input.StaleAfter) {
			continue
		}
		if err := c.saver.Save(ctx, q); err != nil {
			if errors.Is(err, libversioned.ErrConcurrentModification) {
				continue
			}
			return &questionservice.ReclaimStaleAudioOutput{Reclaimed: reclaimed}, fmt.Errorf("save reclaimed question %s: %w", q.ID(), err)
		}

		reclaimed++
	}
	return &questionservice.ReclaimStaleAudioOutput{Reclaimed: reclaimed}, nil
}

// mapAudioSaveError translates the save-layer optimistic-lock error into the
// audio-domain equivalent so handlers can surface 409 Conflict consistently.
func mapAudioSaveError(action string, err error) error {
	if errors.Is(err, libversioned.ErrConcurrentModification) {
		return domain.ErrAudioConcurrentModification
	}
	return fmt.Errorf("%s: %w", action, err)
}

// logSkippedNonWordFill emits a structured warning for an item dropped from
// the pending audio list because its question type does not drive TTS. Today
// only word_fill drives TTS; the gateway's pending-status query is
// type-agnostic, so a future "audio for multiple_choice" rollout could
// backfill non-word_fill items here. Logging the skip surfaces that
// backpressure to operators instead of letting the batch silently report
// processed=0.
func logSkippedNonWordFill(ctx context.Context, q *domainquestion.Question) {
	slog.WarnContext(ctx, "skipping non-word_fill question in pending audio list",
		slog.String("questionId", q.ID()),
		slog.String("workbookId", q.WorkbookID()),
		slog.String("questionType", q.QuestionType().Value()),
	)
}
