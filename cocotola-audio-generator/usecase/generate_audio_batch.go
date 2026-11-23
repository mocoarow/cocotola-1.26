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

// maxFailReasonRunes mirrors the question API's LastError cap.
const maxFailReasonRunes = 500

// ErrUnsupportedLanguage is returned when no TTS voice is configured for a language.
var ErrUnsupportedLanguage = errors.New("unsupported language")

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

	processed, errs := processItems(ctx, logger, api, tts, storage, cfg, items)
	if len(errs) > 0 {
		return processed, fmt.Errorf("%d of %d items failed: %w", len(errs), len(items), errors.Join(errs...))
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return processed, fmt.Errorf("batch interrupted: %w", ctxErr)
	}
	return processed, nil
}

// processItems keeps failures after ctx is done out of errs: main exits 0 on
// any error wrapping context.Canceled, which would hide earlier failures.
func processItems(
	ctx context.Context,
	logger *slog.Logger,
	api QuestionAPI,
	tts TTS,
	storage Storage,
	cfg BatchConfig,
	items []domain.PendingItem,
) (int, []error) {
	processed := 0
	var errs []error
	for _, item := range items {
		processed++
		err := processOne(ctx, logger, api, tts, storage, cfg, item)
		if err == nil {
			continue
		}
		logger.ErrorContext(ctx, "process item failed",
			slog.String("workbookId", item.WorkbookID),
			slog.String("questionId", item.QuestionID),
			slog.Any("error", err),
		)
		if ctx.Err() != nil {
			break
		}
		errs = append(errs, fmt.Errorf("process question %s: %w", item.QuestionID, err))
	}
	return processed, errs
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
		return reportFailure(ctx, logger, api, item, synthErr)
	}
	if err := api.Complete(ctx, item, refs); err != nil {
		if errors.Is(err, domain.ErrAudioStateChanged) {
			logStateChanged(ctx, logger, item)
			return nil
		}
		return fmt.Errorf("complete: %w", err)
	}
	logger.InfoContext(ctx, "audio generated",
		slog.String("workbookId", item.WorkbookID),
		slog.String("questionId", item.QuestionID),
	)
	return nil
}

func reportFailure(
	ctx context.Context,
	logger *slog.Logger,
	api QuestionAPI,
	item domain.PendingItem,
	synthErr error,
) error {
	failErr := api.Fail(ctx, item, truncate(synthErr.Error(), maxFailReasonRunes))
	switch {
	case failErr == nil:
		logger.WarnContext(ctx, "audio generation failed, reported for retry",
			slog.String("workbookId", item.WorkbookID),
			slog.String("questionId", item.QuestionID),
			slog.Any("error", synthErr),
		)
		return nil
	case errors.Is(failErr, domain.ErrAudioStateChanged):
		logStateChanged(ctx, logger, item, slog.Any("error", synthErr))
		return nil
	default:
		// Both errors matter (the synth failure caused the call, the fail
		// call itself failed the cleanup); join so unwrap can see both.
		return fmt.Errorf("synth+fail report: %w", errors.Join(synthErr, failErr))
	}
}

func logStateChanged(ctx context.Context, logger *slog.Logger, item domain.PendingItem, attrs ...slog.Attr) {
	logger.LogAttrs(ctx, slog.LevelInfo, "skipping item changed during generation",
		append([]slog.Attr{
			slog.String("workbookId", item.WorkbookID),
			slog.String("questionId", item.QuestionID),
		}, attrs...)...,
	)
}

func synthesizeAndUpload(
	ctx context.Context,
	tts TTS,
	storage Storage,
	cfg BatchConfig,
	item domain.PendingItem,
) (map[string]domain.AudioRef, error) {
	jobs, err := buildSynthJobs(cfg, item)
	if err != nil {
		return nil, err
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

type synthJob struct {
	slot      string
	text      string
	voice     string
	lang      string
	objectKey string
}

// buildSynthJobs resolves every slot's voice before any TTS call.
func buildSynthJobs(cfg BatchConfig, item domain.PendingItem) ([]synthJob, error) {
	slots := []struct {
		slot string
		text string
		lang string
	}{
		{slot: domain.SlotSource, text: item.SourceText, lang: item.SourceLang},
		{slot: domain.SlotTarget, text: item.TargetText, lang: item.TargetLang},
	}

	jobs := make([]synthJob, 0, len(slots))
	for _, s := range slots {
		voice, fullLang, err := pickVoice(cfg.Voices, s.lang)
		if err != nil {
			return nil, fmt.Errorf("pick %s voice: %w", s.slot, err)
		}
		jobs = append(jobs, synthJob{
			slot:      s.slot,
			text:      s.text,
			voice:     voice,
			lang:      fullLang,
			objectKey: fmt.Sprintf("audio/questions/%s/%s%s", item.QuestionID, s.slot, cfg.ObjectExt),
		})
	}
	return jobs, nil
}

// pickVoice returns the TTS voice name and full locale for a short language code.
func pickVoice(v VoiceConfig, lang string) (string, string, error) {
	switch lang {
	case "ja":
		return v.JaVoice, v.JaLang, nil
	case "en":
		return v.EnVoice, v.EnLang, nil
	default:
		return "", "", fmt.Errorf("%w: %q", ErrUnsupportedLanguage, lang)
	}
}

// truncate returns at most maxRunes runes of s.
func truncate(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes])
}
