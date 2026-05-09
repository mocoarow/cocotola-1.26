package question

import (
	"fmt"
	"time"

	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// markAudioPendingIfWordFill marks the audio generation queue entry of a
// word_fill question as pending when its input text has changed (or when no
// queue entry exists yet). For other question types it is a no-op.
//
// The function prefers the parsed content cached on the aggregate by
// NewQuestion / Edit (which already JSON-decoded the content during
// validation). When the cache is absent (e.g. the aggregate was loaded via
// ReconstructQuestion), it falls back to a fresh parse.
//
// The function relies on Question.MarkAudioPending being a no-op when the
// inputHash is unchanged, so calling it on every save is safe.
func markAudioPendingIfWordFill(q *domainquestion.Question, now time.Time) error {
	if q.QuestionType().Value() != domainquestion.TypeWordFill().Value() {
		return nil
	}
	c := q.WordFillContentParsed()
	if c == nil {
		parsed, err := domainquestion.ParseWordFillContent(q.Content())
		if err != nil {
			return fmt.Errorf("parse word_fill content: %w", err)
		}
		c = parsed
	}
	hash := domainquestion.ComputeWordFillAudioInputHash(*c)
	q.MarkAudioPending(hash, now)
	return nil
}

// toServiceItem projects a domain Question into the service-layer Item DTO,
// including audio refs when ready.
func toServiceItem(q *domainquestion.Question) questionservice.Item {
	return questionservice.Item{
		QuestionID:   q.ID(),
		QuestionType: q.QuestionType().Value(),
		Content:      q.Content(),
		Tags:         q.Tags(),
		OrderIndex:   q.OrderIndex(),
		CreatedAt:    q.CreatedAt(),
		UpdatedAt:    q.UpdatedAt(),
		Audio:        audioOutputFromQuestion(q),
	}
}

// audioOutputFromQuestion projects a question's ready audio refs into the
// service-layer DTO. Returns nil when the audio is not yet ready.
func audioOutputFromQuestion(q *domainquestion.Question) *questionservice.AudioOutput {
	ag := q.AudioGeneration()
	if ag == nil || !ag.Status().IsReady() {
		return nil
	}
	refs := ag.Refs()
	out := &questionservice.AudioOutput{
		Source: nil,
		Target: nil,
	}
	if ref, ok := refs[domainquestion.AudioLangSource]; ok {
		out.Source = &questionservice.AudioRefOutput{
			Path:        ref.Path(),
			DurationSec: ref.DurationSec(),
		}
	}
	if ref, ok := refs[domainquestion.AudioLangTarget]; ok {
		out.Target = &questionservice.AudioRefOutput{
			Path:        ref.Path(),
			DurationSec: ref.DurationSec(),
		}
	}
	if out.Source == nil && out.Target == nil {
		return nil
	}
	return out
}
