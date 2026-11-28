// Package main is the entry point for the cocotola-audio-generator batch.
//
// The process is designed for one-shot execution by Cloud Run Jobs (or any
// scheduler that re-runs it periodically). It pulls pending audio items from
// cocotola-question, synthesizes audio with Google Cloud Text-to-Speech, and
// uploads the resulting files to Google Cloud Storage.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/usecase"
)

// runTimeout caps each scheduled run. Cloud Run Jobs has its own task timeout
// but we want a process-internal limit so a stuck Cloud TTS call doesn't keep
// the whole batch container alive.
const runTimeout = 30 * time.Minute

const (
	jaLanguageCode = "ja-JP"
	enLanguageCode = "en-US"
)

func main() {
	exitCode, err := run()
	if err != nil {
		slog.Error("audio generator run", slog.Any("error", err))
	}
	os.Exit(exitCode)
}

func run() (int, error) {
	// Honor SIGTERM (Cloud Run sends it ~10s before forced kill on preemption)
	// and SIGINT so in-flight TTS / GCS calls can be canceled cleanly instead
	// of hard-killed.
	base, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	ctx, cancel := context.WithTimeout(base, runTimeout)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		return 1, fmt.Errorf("load config: %w", err)
	}

	logger := slog.Default().With(
		slog.String("app", "cocotola-audio-generator"),
		slog.String("env", cfg.AppEnv),
	)

	if err := generate(ctx, logger, cfg.Audio); err != nil {
		return 1, err
	}
	return 0, nil
}

func generate(ctx context.Context, logger *slog.Logger, cfg config.AudioGeneratorConfig) error {
	enc, err := gateway.ParseAudioEncoding(cfg.AudioEncoding)
	if err != nil {
		return fmt.Errorf("parse audio encoding: %w", err)
	}

	tts, err := gateway.NewTTSClient(ctx, enc, cfg.SampleRateHz)
	if err != nil {
		return fmt.Errorf("new tts client: %w", err)
	}
	defer func() {
		if err := tts.Close(); err != nil {
			logger.Error("close tts client", slog.Any("error", err))
		}
	}()

	gcs, err := gateway.NewGCSClient(ctx, cfg.BucketName)
	if err != nil {
		return fmt.Errorf("new gcs client: %w", err)
	}
	defer func() {
		if err := gcs.Close(); err != nil {
			logger.Error("close gcs client", slog.Any("error", err))
		}
	}()

	api := gateway.NewQuestionAPIClient(
		cfg.QuestionAPIBaseURL,
		cfg.QuestionAPIKey,
		time.Duration(cfg.QuestionAPITimeoutSec)*time.Second,
	)

	processed, err := usecase.GenerateAudioBatch(ctx, logger, api, tts, gcs, newBatchConfig(cfg, enc))
	logger.InfoContext(ctx, "batch finished", slog.Int("processed", processed))
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("generate audio batch: %w", err)
	}
	return nil
}

func newBatchConfig(cfg config.AudioGeneratorConfig, enc gateway.AudioEncoding) usecase.BatchConfig {
	return usecase.BatchConfig{
		MaxPerRun:   cfg.MaxPerRun,
		ContentType: enc.ContentType,
		ObjectExt:   enc.ObjectExt,
		Voices: usecase.VoiceConfig{
			JaVoice: cfg.TTSVoiceJa,
			JaLang:  jaLanguageCode,
			EnVoice: cfg.TTSVoiceEn,
			EnLang:  enLanguageCode,
		},
		StaleAfter: time.Duration(cfg.StaleAfterSec) * time.Second,
	}
}
