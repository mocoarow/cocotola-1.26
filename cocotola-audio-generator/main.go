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

	enc, err := gateway.ParseAudioEncoding(cfg.Audio.AudioEncoding)
	if err != nil {
		return 1, fmt.Errorf("parse audio encoding: %w", err)
	}

	tts, err := gateway.NewTTSClient(ctx, cfg.Audio.AudioEncoding, cfg.Audio.SampleRateHz)
	if err != nil {
		return 1, fmt.Errorf("new tts client: %w", err)
	}
	defer func() {
		if err := tts.Close(); err != nil {
			logger.Error("close tts client", slog.Any("error", err))
		}
	}()

	gcs, err := gateway.NewGCSClient(ctx, cfg.Audio.BucketName)
	if err != nil {
		return 1, fmt.Errorf("new gcs client: %w", err)
	}
	defer func() {
		if err := gcs.Close(); err != nil {
			logger.Error("close gcs client", slog.Any("error", err))
		}
	}()

	api := gateway.NewQuestionAPIClient(
		cfg.Audio.QuestionAPIBaseURL,
		cfg.Audio.QuestionAPIKey,
		time.Duration(cfg.Audio.QuestionAPITimeoutSec)*time.Second,
	)

	processed, err := usecase.GenerateAudioBatch(
		ctx,
		logger,
		api,
		tts,
		gcs,
		usecase.BatchConfig{
			MaxPerRun:   cfg.Audio.MaxPerRun,
			ContentType: enc.ContentType,
			ObjectExt:   enc.ObjectExt,
			Voices: usecase.VoiceConfig{
				JaVoice: cfg.Audio.TTSVoiceJa,
				JaLang:  "ja-JP",
				EnVoice: cfg.Audio.TTSVoiceEn,
				EnLang:  "en-US",
			},
			StaleAfter: time.Duration(cfg.Audio.StaleAfterSec) * time.Second,
		},
	)
	logger.InfoContext(ctx, "batch finished", slog.Int("processed", processed))
	if err != nil && !errors.Is(err, context.Canceled) {
		return 1, fmt.Errorf("generate audio batch: %w", err)
	}
	return 0, nil
}

