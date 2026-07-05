package gateway

import (
	"fmt"

	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// AudioEncoding bundles all per-format constants so that adding a new encoding
// requires changes only in this file.
type AudioEncoding struct {
	Name        string
	ContentType string
	ObjectExt   string
	ttsEncoding texttospeechpb.AudioEncoding
}

// TTSEncoding returns the Cloud TTS protobuf encoding value.
func (e AudioEncoding) TTSEncoding() texttospeechpb.AudioEncoding {
	return e.ttsEncoding
}

// ParseAudioEncoding looks up the AudioEncoding for the given name.
func ParseAudioEncoding(name string) (AudioEncoding, error) {
	switch name {
	case "OGG_OPUS":
		return AudioEncoding{
			Name:        "OGG_OPUS",
			ContentType: "audio/ogg; codecs=opus",
			ObjectExt:   ".opus",
			ttsEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
		}, nil
	case "MP3":
		return AudioEncoding{
			Name:        "MP3",
			ContentType: "audio/mpeg",
			ObjectExt:   ".mp3",
			ttsEncoding: texttospeechpb.AudioEncoding_MP3,
		}, nil
	default:
		return AudioEncoding{}, fmt.Errorf("unsupported audio encoding %q", name)
	}
}
