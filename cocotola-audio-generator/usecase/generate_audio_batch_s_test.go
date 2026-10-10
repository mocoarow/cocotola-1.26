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

func Test_truncate_shouldReturnInputUnchanged_whenWithinLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxRunes int
	}{
		{name: "empty string", input: "", maxRunes: 3},
		{name: "ascii shorter than limit", input: "ab", maxRunes: 3},
		{name: "ascii exactly at limit", input: "abc", maxRunes: 3},
		{name: "multibyte exactly at limit", input: "りんご", maxRunes: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: input whose rune count does not exceed maxRunes

			// when
			got := usecase.Truncate(tt.input, tt.maxRunes)

			// then
			assert.Equal(t, tt.input, got)
		})
	}
}

func Test_truncate_shouldCutAtRuneBoundary_whenExceedingLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxRunes int
		want     string
	}{
		{name: "ascii", input: "hello world", maxRunes: 5, want: "hello"},
		{name: "ascii one over limit", input: "abcd", maxRunes: 3, want: "abc"},
		{name: "multibyte", input: "りんごを食べる", maxRunes: 3, want: "りんご"},
		{name: "multibyte one over limit", input: "りんごを", maxRunes: 3, want: "りんご"},
		{name: "mixed ascii and multibyte", input: "aりbん", maxRunes: 2, want: "aり"},
		{name: "zero limit", input: "abc", maxRunes: 0, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: input whose rune count exceeds maxRunes

			// when
			got := usecase.Truncate(tt.input, tt.maxRunes)

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_pickVoice_shouldReturnConfiguredVoice_whenLangIsSupported(t *testing.T) {
	t.Parallel()

	voices := defaultBatchConfig().Voices
	tests := []struct {
		lang string
		want string
	}{
		{lang: "ja", want: voices.JaVoice},
		{lang: "en", want: voices.EnVoice},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()

			// given: a supported short language code

			// when
			got := usecase.PickVoice(voices, tt.lang)

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_pickVoice_shouldReturnEmpty_whenLangIsUnsupported(t *testing.T) {
	t.Parallel()

	voices := defaultBatchConfig().Voices
	tests := []struct {
		name string
		lang string
	}{
		{name: "unknown code", lang: "fr"},
		{name: "empty code", lang: ""},
		{name: "uppercase code", lang: "JA"},
		{name: "full locale", lang: "ja-JP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: a language code that is not a supported short code

			// when
			got := usecase.PickVoice(voices, tt.lang)

			// then
			assert.Empty(t, got)
		})
	}
}

func Test_pickFullLang_shouldReturnConfiguredLocale_whenLangIsSupported(t *testing.T) {
	t.Parallel()

	voices := defaultBatchConfig().Voices
	tests := []struct {
		lang string
		want string
	}{
		{lang: "ja", want: voices.JaLang},
		{lang: "en", want: voices.EnLang},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()

			// given: a supported short language code

			// when
			got := usecase.PickFullLang(voices, tt.lang)

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_pickFullLang_shouldReturnEmpty_whenLangIsUnsupported(t *testing.T) {
	t.Parallel()

	voices := defaultBatchConfig().Voices
	tests := []struct {
		name string
		lang string
	}{
		{name: "unknown code", lang: "fr"},
		{name: "empty code", lang: ""},
		{name: "uppercase code", lang: "JA"},
		{name: "full locale", lang: "ja-JP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: a language code that is not a supported short code

			// when
			got := usecase.PickFullLang(voices, tt.lang)

			// then
			assert.Empty(t, got)
		})
	}
}
