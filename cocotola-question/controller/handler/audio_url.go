package handler

import (
	"strings"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// AudioURLBuilder turns bucket-relative audio paths into public URLs for API responses.
type AudioURLBuilder struct {
	baseURL string
}

// NewAudioURLBuilder returns a builder that joins audio paths with the given public base URL.
func NewAudioURLBuilder(baseURL string) *AudioURLBuilder {
	return &AudioURLBuilder{baseURL: strings.TrimRight(baseURL, "/")}
}

// QuestionAudio converts service-layer audio refs into the wire shape; a slot
// without audio gets an empty Url, per the OpenAPI schema.
func (b *AudioURLBuilder) QuestionAudio(audio *questionservice.AudioOutput) api.QuestionAudio {
	if audio == nil {
		return api.QuestionAudio{Source: emptyAudioRef(), Target: emptyAudioRef()}
	}
	return api.QuestionAudio{
		Source: b.audioRef(audio.Source),
		Target: b.audioRef(audio.Target),
	}
}

func (b *AudioURLBuilder) audioRef(ref *questionservice.AudioRefOutput) api.QuestionAudioRef {
	if ref == nil {
		return emptyAudioRef()
	}
	return api.QuestionAudioRef{Url: b.urlFor(ref.Path), DurationSec: ref.DurationSec}
}

func emptyAudioRef() api.QuestionAudioRef {
	return api.QuestionAudioRef{Url: "", DurationSec: 0}
}

func (b *AudioURLBuilder) urlFor(path string) string {
	if path == "" {
		return ""
	}
	return b.baseURL + "/" + strings.TrimLeft(path, "/")
}
