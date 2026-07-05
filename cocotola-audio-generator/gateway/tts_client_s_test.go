package gateway_test

import (
	"testing"

	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/gateway"
)

func Test_ParseAudioEncoding_shouldReturnOggOpus_whenOGG_OPUSGiven(t *testing.T) {
	t.Parallel()

	// when
	enc, err := gateway.ParseAudioEncoding("OGG_OPUS")

	// then
	require.NoError(t, err)
	assert.Equal(t, "OGG_OPUS", enc.Name)
	assert.Equal(t, "audio/ogg; codecs=opus", enc.ContentType)
	assert.Equal(t, ".opus", enc.ObjectExt)
	assert.Equal(t, texttospeechpb.AudioEncoding_OGG_OPUS, enc.TTSEncoding())
}

func Test_ParseAudioEncoding_shouldReturnMP3_whenMP3Given(t *testing.T) {
	t.Parallel()

	// when
	enc, err := gateway.ParseAudioEncoding("MP3")

	// then
	require.NoError(t, err)
	assert.Equal(t, "MP3", enc.Name)
	assert.Equal(t, "audio/mpeg", enc.ContentType)
	assert.Equal(t, ".mp3", enc.ObjectExt)
	assert.Equal(t, texttospeechpb.AudioEncoding_MP3, enc.TTSEncoding())
}

func Test_ParseAudioEncoding_shouldReturnError_whenUnknownEncodingGiven(t *testing.T) {
	t.Parallel()

	// when
	enc, err := gateway.ParseAudioEncoding("FLAC")

	// then
	require.ErrorContains(t, err, "FLAC")
	assert.Equal(t, gateway.AudioEncoding{}, enc)
}
