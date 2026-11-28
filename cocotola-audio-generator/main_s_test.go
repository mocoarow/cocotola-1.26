package main_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/mocoarow/cocotola-1.26/cocotola-audio-generator"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/usecase"
)

func Test_newBatchConfig_shouldMapAudioConfig_whenEncodingIsParsed(t *testing.T) {
	t.Parallel()

	// given
	cfg := config.AudioGeneratorConfig{
		BucketName:            "test-bucket",
		MaxPerRun:             37,
		TTSVoiceJa:            "ja-JP-Wavenet-D",
		TTSVoiceEn:            "en-US-Wavenet-F",
		AudioEncoding:         "MP3",
		SampleRateHz:          16000,
		QuestionAPIBaseURL:    "http://question.test:9999",
		QuestionAPIKey:        "test-key",
		QuestionAPITimeoutSec: 45,
		StaleAfterSec:         120,
	}
	enc, err := gateway.ParseAudioEncoding(cfg.AudioEncoding)
	require.NoError(t, err)
	expected := usecase.BatchConfig{
		MaxPerRun:   37,
		ContentType: "audio/mpeg",
		ObjectExt:   ".mp3",
		Voices: usecase.VoiceConfig{
			JaVoice: "ja-JP-Wavenet-D",
			JaLang:  "ja-JP",
			EnVoice: "en-US-Wavenet-F",
			EnLang:  "en-US",
		},
		StaleAfter: 2 * time.Minute,
	}

	// when
	got := main.NewBatchConfig(cfg, enc)

	// then
	assert.Equal(t, expected, got)
}
