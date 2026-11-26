package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/usecase"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/audio"
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

func failReasonFor(slot string, cause error) any {
	return mock.MatchedBy(func(reason string) bool {
		return strings.Contains(reason, slot) && strings.Contains(reason, cause.Error())
	})
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

func Test_GenerateAudioBatch_shouldFailItem_whenSourceUploadFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: no expectation for the target slot or Complete
	item := sampleItem()
	sourceAudio := []byte("source-audio")
	uploadErr := errors.New("storage unavailable")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Fail(ctx, item, failReasonFor(domain.SlotSource, uploadErr)).Return(nil).Once()
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, item.SourceText, mock.Anything, mock.Anything).Return(sourceAudio, nil).Once()
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, sourceAudio).Return(int64(0), uploadErr).Once()

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
}

func Test_GenerateAudioBatch_shouldFailItem_whenTargetSynthesisFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: no expectation for the target upload or Complete
	item := sampleItem()
	sourceAudio := []byte("source-audio")
	synthErr := errors.New("tts quota exceeded")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Fail(ctx, item, failReasonFor(domain.SlotTarget, synthErr)).Return(nil).Once()
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, item.SourceText, mock.Anything, mock.Anything).Return(sourceAudio, nil).Once()
	tts.EXPECT().Synthesize(ctx, item.TargetText, mock.Anything, mock.Anything).Return(nil, synthErr).Once()
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, sourceAudio).Return(int64(5), nil).Once()

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
}

func Test_GenerateAudioBatch_shouldFailItem_whenTargetUploadFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: no expectation for Complete
	item := sampleItem()
	sourceAudio := []byte("source-audio")
	targetAudio := []byte("target-audio")
	uploadErr := errors.New("storage unavailable")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Fail(ctx, item, failReasonFor(domain.SlotTarget, uploadErr)).Return(nil).Once()
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, item.SourceText, mock.Anything, mock.Anything).Return(sourceAudio, nil).Once()
	tts.EXPECT().Synthesize(ctx, item.TargetText, mock.Anything, mock.Anything).Return(targetAudio, nil).Once()
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, sourceAudio).Return(int64(5), nil).Once()
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, targetAudio).Return(int64(0), uploadErr).Once()

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
}

func Test_GenerateAudioBatch_shouldFailItemWithoutSynthesis_whenLangIsUnsupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sourceLang string
		targetLang string
	}{
		{name: "unsupported source", sourceLang: "fr", targetLang: "en"},
		{name: "unsupported target", sourceLang: "ja", targetLang: "fr"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given: TTS and Storage mocks reject any call
			item := sampleItem()
			item.SourceLang = tt.sourceLang
			item.TargetLang = tt.targetLang
			api := NewMockQuestionAPI(t)
			api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
			api.EXPECT().Claim(ctx, item).Return(nil)
			api.EXPECT().Fail(ctx, item, mock.MatchedBy(func(reason string) bool {
				return strings.Contains(reason, usecase.ErrUnsupportedLanguage.Error())
			})).Return(nil)

			// when
			_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, NewMockTTS(t), NewMockStorage(t), defaultBatchConfig())

			// then: the item is failed with the unsupported-language reason
			require.NoError(t, err)
		})
	}
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

func Test_GenerateAudioBatch_shouldReturnError_whenClaimFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	claimErr := errors.New("claim transport")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(claimErr)
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.ErrorIs(t, err, claimErr)
}

func Test_GenerateAudioBatch_shouldReturnError_whenCompleteFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	completeErr := errors.New("complete transport")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Complete(ctx, item, mock.Anything).Return(completeErr)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return([]byte("audio"), nil).Times(2)
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil).Times(2)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.ErrorIs(t, err, completeErr)
}

func Test_GenerateAudioBatch_shouldReturnError_whenFailReportFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	failErr := errors.New("fail transport")
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Fail(ctx, item, mock.Anything).Return(failErr)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("synth boom"))
	storage := NewMockStorage(t)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.ErrorIs(t, err, failErr)
}

func Test_GenerateAudioBatch_shouldProcessRemainingItems_whenEarlierItemFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	first := sampleItem()
	second := sampleItem()
	second.QuestionID = "q-2"
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{first, second}, nil)
	api.EXPECT().Claim(ctx, first).Return(errors.New("claim transport"))
	api.EXPECT().Claim(ctx, second).Return(nil)
	api.EXPECT().Complete(ctx, second, mock.Anything).Return(nil)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return([]byte("audio"), nil).Times(2)
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil).Times(2)

	// when
	_, _ = usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	api.AssertCalled(t, "Complete", ctx, second, mock.Anything)
}

func Test_GenerateAudioBatch_shouldStopProcessing_whenContextIsCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// given
	first := sampleItem()
	second := sampleItem()
	second.QuestionID = "q-2"
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{first, second}, nil)
	api.EXPECT().Claim(ctx, first).RunAndReturn(func(context.Context, domain.PendingItem) error {
		cancel()
		return fmt.Errorf("do claim request: %w", context.Canceled)
	})
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	_, _ = usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	api.AssertNotCalled(t, "Claim", ctx, second)
}

func Test_GenerateAudioBatch_shouldReturnCanceled_whenOnlyCancellationInterruptsBatch(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// given
	item := sampleItem()
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).RunAndReturn(func(context.Context, domain.PendingItem) error {
		cancel()
		return fmt.Errorf("do claim request: %w", context.Canceled)
	})
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.ErrorIs(t, err, context.Canceled)
}

func Test_GenerateAudioBatch_shouldNotReportCanceled_whenEarlierItemFailedBeforeCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// given
	first := sampleItem()
	second := sampleItem()
	second.QuestionID = "q-2"
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{first, second}, nil)
	claimErr := errors.New("claim transport")
	api.EXPECT().Claim(ctx, first).Return(claimErr)
	api.EXPECT().Claim(ctx, second).RunAndReturn(func(context.Context, domain.PendingItem) error {
		cancel()
		return fmt.Errorf("do claim request: %w", context.Canceled)
	})
	tts := NewMockTTS(t)
	storage := NewMockStorage(t)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then: main exits 0 for context.Canceled, which would hide the first failure
	require.ErrorIs(t, err, claimErr)
	assert.NotErrorIs(t, err, context.Canceled)
}

func Test_GenerateAudioBatch_shouldSkipItem_whenCompleteReportsStateChanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Complete(ctx, item, mock.Anything).Return(domain.ErrAudioStateChanged)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return([]byte("audio"), nil).Times(2)
	storage := NewMockStorage(t)
	storage.EXPECT().Upload(ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil).Times(2)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
}

func Test_GenerateAudioBatch_shouldSkipItem_whenFailReportsStateChanged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := sampleItem()
	api := NewMockQuestionAPI(t)
	api.EXPECT().ListPending(ctx, mock.Anything).Return([]domain.PendingItem{item}, nil)
	api.EXPECT().Claim(ctx, item).Return(nil)
	api.EXPECT().Fail(ctx, item, mock.Anything).Return(domain.ErrAudioStateChanged)
	tts := NewMockTTS(t)
	tts.EXPECT().Synthesize(ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("synth boom"))
	storage := NewMockStorage(t)

	// when
	_, err := usecase.GenerateAudioBatch(ctx, newDiscardLogger(), api, tts, storage, defaultBatchConfig())

	// then
	require.NoError(t, err)
}

func Test_GenerateAudioBatch_shouldCallReclaimStale_whenStaleAfterIsConfigured(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	cfg := defaultBatchConfig()
	cfg.StaleAfter = 15 * time.Minute
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
	cfg.StaleAfter = 15 * time.Minute
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

func Test_pickVoice_shouldReturnConfiguredVoiceAndLocale_whenLangIsSupported(t *testing.T) {
	t.Parallel()

	voices := defaultBatchConfig().Voices
	tests := []struct {
		lang      string
		wantVoice string
		wantLang  string
	}{
		{lang: "ja", wantVoice: voices.JaVoice, wantLang: voices.JaLang},
		{lang: "en", wantVoice: voices.EnVoice, wantLang: voices.EnLang},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()

			// given: a supported short language code

			// when
			gotVoice, gotLang, err := usecase.PickVoice(voices, tt.lang)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.wantVoice, gotVoice)
			assert.Equal(t, tt.wantLang, gotLang)
		})
	}
}

func Test_pickVoice_shouldResolveVoice_whenLangIsInSharedSupportedLangs(t *testing.T) {
	t.Parallel()

	voices := defaultBatchConfig().Voices
	for _, lang := range audio.SupportedLangs() {
		t.Run(lang, func(t *testing.T) {
			t.Parallel()

			// given: a language cocotola-question queues audio for

			// when
			_, _, err := usecase.PickVoice(voices, lang)

			// then
			require.NoError(t, err)
		})
	}
}

func Test_pickVoice_shouldReturnErrUnsupportedLanguage_whenLangIsUnsupported(t *testing.T) {
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
			_, _, err := usecase.PickVoice(voices, tt.lang)

			// then
			require.ErrorIs(t, err, usecase.ErrUnsupportedLanguage)
		})
	}
}
