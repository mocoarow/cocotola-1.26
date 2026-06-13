package gateway_test

import (
	"testing"

	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/gateway"
)

func Test_parseEncoding_shouldReturnOggOpus_whenOGG_OPUSGiven(t *testing.T) {
	t.Parallel()

	// when
	enc, err := gateway.ParseEncoding("OGG_OPUS")

	// then
	require.NoError(t, err)
	assert.Equal(t, texttospeechpb.AudioEncoding_OGG_OPUS, enc)
}

func Test_parseEncoding_shouldReturnMP3_whenMP3Given(t *testing.T) {
	t.Parallel()

	// when
	enc, err := gateway.ParseEncoding("MP3")

	// then
	require.NoError(t, err)
	assert.Equal(t, texttospeechpb.AudioEncoding_MP3, enc)
}

func Test_parseEncoding_shouldReturnError_whenUnknownEncodingGiven(t *testing.T) {
	t.Parallel()

	// when
	enc, err := gateway.ParseEncoding("FLAC")

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FLAC")
	assert.Equal(t, texttospeechpb.AudioEncoding_AUDIO_ENCODING_UNSPECIFIED, enc)
}
