// Package usecase orchestrates the audio generation batch.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/domain"
)

// QuestionAPI is the subset of cocotola-question's internal API the batch needs.
type QuestionAPI interface {
	ReclaimStale(ctx context.Context, staleAfter time.Duration, limit int) (int, error)
	ListPending(ctx context.Context, limit int) ([]domain.PendingItem, error)
	Claim(ctx context.Context, item domain.PendingItem) error
	Complete(ctx context.Context, item domain.PendingItem, refs map[string]domain.AudioRef) error
	Fail(ctx context.Context, item domain.PendingItem, reason string) error
}

// TTS is the subset of TTS functionality the batch needs.
type TTS interface {
	Synthesize(ctx context.Context, text, voice, lang string) ([]byte, error)
}

// Storage is the subset of object-storage functionality the batch needs.
type Storage interface {
	Upload(ctx context.Context, objectPath, contentType string, data []byte) (int64, error)
}

// VoiceConfig maps a BCP-47 short language code (e.g. "en", "ja") to a
// concrete TTS voice name and full BCP-47 locale.
type VoiceConfig struct {
	JaVoice string
	JaLang  string
	EnVoice string
	EnLang  string
}

// maxFailReasonRunes caps the failure reason sent to the question API so a
// long Cloud TTS error stack-trace does not blow past the server's stored
// LastError limit. The server applies the same cap; truncating client-side
// keeps the request body small.
const maxFailReasonRunes = 500

// BatchConfig collects runtime knobs for one batch run.
type BatchConfig struct {
	MaxPerRun   int
	ContentType string // e.g. "audio/ogg; codecs=opus"
	ObjectExt   string // e.g. ".opus"
	Voices      VoiceConfig
	// StaleAfter is how long an audio entry may sit in "generating" state
	// before the reaper at the start of the batch reclaims it back to
	// "pending". Set to 0 to skip the reclaim sweep.
	StaleAfter time.Duration
}

// GenerateAudioBatch runs a single pass: fetch pending items, claim, synthesize,
// upload, and report success/failure back to cocotola-question.
//
// It does not retry within a single run — failed items are marked failed and
// retried on the next scheduled run with backoff applied via failedAttempts.
func GenerateAudioBatch(
	ctx context.Context,
	logger *slog.Logger,
	api QuestionAPI,
	tts TTS,
	storage Storage,
	cfg BatchConfig,
) (processed int, err error) {
	if cfg.StaleAfter > 0 {
		reclaimed, reclaimErr := api.ReclaimStale(ctx, cfg.StaleAfter, cfg.MaxPerRun)
		if reclaimErr != nil {
			// A reclaim failure is logged but does not abort the run: pending
			// items can still be processed; the reclaim sweep will retry next
			// run.
			logger.WarnContext(ctx, "reclaim stale failed, continuing",
				slog.Any("error", reclaimErr))
		} else if reclaimed > 0 {
			logger.InfoContext(ctx, "reclaimed stuck items",
				slog.Int("reclaimed", reclaimed))
		}
	}
	items, err := api.ListPending(ctx, cfg.MaxPerRun)
	if err != nil {
		return 0, fmt.Errorf("list pending: %w", err)
	}
	if len(items) == 0 {
		logger.InfoContext(ctx, "no pending audio items")
		return 0, nil
	}
	logger.InfoContext(ctx, "starting audio batch", slog.Int("items", len(items)))

	for _, item := range items {
		processed++
		if err := processOne(ctx, logger, api, tts, storage, cfg, item); err != nil {
			logger.ErrorContext(ctx, "process item failed",
				slog.String("workbookId", item.WorkbookID),
				slog.String("questionId", item.QuestionID),
				slog.Any("error", err),
			)
			// processOne already calls api.Fail on synth/upload errors; we
			// only land here for transport-level failures, which we let bubble
			// only after attempting all items.
		}
	}
	return processed, nil
}

func processOne(
	ctx context.Context,
	logger *slog.Logger,
	api QuestionAPI,
	tts TTS,
	storage Storage,
	cfg BatchConfig,
	item domain.PendingItem,
) error {
	if err := api.Claim(ctx, item); err != nil {
		if errors.Is(err, domain.ErrClaimRace) {
			logger.InfoContext(ctx, "skipping item already claimed",
				slog.String("questionId", item.QuestionID))
			return nil
		}
		return fmt.Errorf("claim: %w", err)
	}

	refs, synthErr := synthesizeAndUpload(ctx, tts, storage, cfg, item)
	if synthErr != nil {
		if failErr := api.Fail(ctx, item, truncate(synthErr.Error(), maxFailReasonRunes)); failErr != nil {
			// Both errors matter (the synth failure caused the call, the fail
			// call itself failed the cleanup); join so unwrap can see both.
			return fmt.Errorf("synth+fail report: %w", errors.Join(synthErr, failErr))
		}
		return fmt.Errorf("synthesize/upload: %w", synthErr)
	}
	if err := api.Complete(ctx, item, refs); err != nil {
		return fmt.Errorf("complete: %w", err)
	}
	logger.InfoContext(ctx, "audio generated",
		slog.String("workbookId", item.WorkbookID),
		slog.String("questionId", item.QuestionID),
	)
	return nil
}

func synthesizeAndUpload(
	ctx context.Context,
	tts TTS,
	storage Storage,
	cfg BatchConfig,
	item domain.PendingItem,
) (map[string]domain.AudioRef, error) {
	jobs := []struct {
		slot      string
		text      string
		voice     string
		lang      string
		objectKey string
	}{
		{
			slot:      domain.SlotSource,
			text:      item.SourceText,
			voice:     pickVoice(cfg.Voices, item.SourceLang),
			lang:      pickFullLang(cfg.Voices, item.SourceLang),
			objectKey: fmt.Sprintf("audio/questions/%s/%s%s", item.QuestionID, domain.SlotSource, cfg.ObjectExt),
		},
		{
			slot:      domain.SlotTarget,
			text:      item.TargetText,
			voice:     pickVoice(cfg.Voices, item.TargetLang),
			lang:      pickFullLang(cfg.Voices, item.TargetLang),
			objectKey: fmt.Sprintf("audio/questions/%s/%s%s", item.QuestionID, domain.SlotTarget, cfg.ObjectExt),
		},
	}

	out := make(map[string]domain.AudioRef, len(jobs))
	for _, job := range jobs {
		audio, err := tts.Synthesize(ctx, job.text, job.voice, job.lang)
		if err != nil {
			return nil, fmt.Errorf("synth %s: %w", job.slot, err)
		}
		size, err := storage.Upload(ctx, job.objectKey, cfg.ContentType, audio)
		if err != nil {
			return nil, fmt.Errorf("upload %s: %w", job.slot, err)
		}
		out[job.slot] = domain.AudioRef{
			Path:        job.objectKey,
			DurationSec: 0, // Cloud TTS does not return duration; left for future calc
			SizeBytes:   size,
		}
	}
	return out, nil
}

func pickVoice(v VoiceConfig, lang string) string {
	switch lang {
	case "ja":
		return v.JaVoice
	case "en":
		return v.EnVoice
	default:
		// fall through; downstream API will reject unsupported voices
		return ""
	}
}

func pickFullLang(v VoiceConfig, lang string) string {
	switch lang {
	case "ja":
		return v.JaLang
	case "en":
		return v.EnLang
	default:
		return ""
	}
}

// truncate returns at most maxRunes runes of s. The argument is rune count,
// not byte count, so multi-byte UTF-8 characters are not split mid-codepoint
// when the failure reason carries non-ASCII text (e.g. error messages from a
// Japanese-language Cloud API response).
func truncate(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes])
}
