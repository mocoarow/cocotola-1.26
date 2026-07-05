package usecase_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/usecase"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func defaultBatchConfig() usecase.BatchConfig {
	return usecase.BatchConfig{
		MaxPerRun:   10,
		ContentType: "audio/ogg; codecs=opus",
		ObjectExt:   ".opus",
		Voices: usecase.VoiceConfig{
			JaVoice: "ja-JP-Neural2-B",
			JaLang:  "ja-JP",
			EnVoice: "en-US-Neural2-C",
			EnLang:  "en-US",
		},
	}
}

func sampleItem() domain.PendingItem {
	return domain.PendingItem{
		WorkbookID: "wb-1",
		QuestionID: "q-1",
		SourceText: "りんごを食べる",
		SourceLang: "ja",
		TargetText: "eat an apple",
		TargetLang: "en",
		InputHash:  "h1",
	}
}

func Test_GenerateAudioBatch_shouldReturnZero_whenNoPending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return(nil, nil)
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	processed, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, processed)
}

func Test_GenerateAudioBatch_shouldClaimAndCompleteItem_whenSynthesisSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Complete(ctx, item, mock.MatchedBy(func(refs map[string]domain.AudioRef) bool {
		_, hasSource := refs[domain.SlotSource]
		_, hasTarget := refs[domain.SlotTarget]
		return hasSource && hasTarget
	})).Return(nil)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return([]byte("audio"), nil).Times(2)
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil).Times(2)

	// when
	processed, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func Test_GenerateAudioBatch_shouldFailItem_whenSynthesisFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	synthErr := errors.New("synth boom")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Fail(ctx, item, mock.Anything).Return(nil)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, synthErr)
	storage := NewMockStorage(t)

	// when
	processed, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then: fail is reported via api.Fail (not via the return error)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func Test_GenerateAudioBatch_shouldSkipItem_whenClaimRaceLost(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(domain.ErrClaimRace)
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	processed, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then: claim race results in skip (no Complete/Fail call made — assert via mock expectations)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

func Test_GenerateAudioBatch_shouldCallReclaimStale_whenStaleAfterIsConfigured(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	cfg := defaultBatchConfig()
	cfg.StaleAfter = 15 * 60 * 1_000_000_000 // 15 minutes in nanoseconds
	api := NewMockQuestionAPI(t)
	api.EXPECT().ReclaimStale(ctx, cfg.StaleAfter, cfg.MaxPerRun).Return(2, nil)
	api.EXPECT().ListPending(ctx, mock.Anything).Return(nil, nil)
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	processed, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, cfg)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, processed)
}

func Test_GenerateAudioBatch_shouldContinueProcessing_whenReclaimStaleFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	cfg := defaultBatchConfig()
	cfg.StaleAfter = 15 * 60 * 1_000_000_000
	api := NewMockQuestionAPI(t)
	api.EXPECT().ReclaimStale(ctx, cfg.StaleAfter, cfg.MaxPerRun).Return(0, errors.New("network down"))
	api.EXPECT().ListPending(ctx, mock.Anything).Return(nil, nil)
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	processed, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, cfg)

	// then: reclaim error logged but not returned
	require.NoError(t, err)
	assert.Equal(t, 0, processed)
}
