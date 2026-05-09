package study

import (
	"strings"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// AudioURLBuilder turns service-layer audio refs (with bucket-relative paths)
// into wire-level api.QuestionAudio values whose URLs include the configured
// public base URL.
type AudioURLBuilder struct {
	baseURL string
}

// NewAudioURLBuilder returns a builder that joins audio paths with the given base URL.
func NewAudioURLBuilder(baseURL string) *AudioURLBuilder {
	return &AudioURLBuilder{baseURL: strings.TrimRight(baseURL, "/")}
}

// audioForStudy converts a study-layer audio object into the wire response shape.
// Empty URLs are preserved (the frontend treats them as "audio not available").
func (b *AudioURLBuilder) audioForStudy(audio *studyservice.QuestionItemAudio) api.QuestionAudio {
	out := api.QuestionAudio{
		Source: emptyAudioRef(),
		Target: emptyAudioRef(),
	}
	if audio == nil {
		return out
	}
	if audio.Source != nil {
		out.Source = api.QuestionAudioRef{
			Url:         b.urlFor(audio.Source.Path),
			DurationSec: audio.Source.DurationSec,
		}
	}
	if audio.Target != nil {
		out.Target = api.QuestionAudioRef{
			Url:         b.urlFor(audio.Target.Path),
			DurationSec: audio.Target.DurationSec,
		}
	}
	return out
}

// emptyAudioRef is the placeholder used when no audio file is available for a
// given slot. See the analogous helper in handler/question for the rationale.
func emptyAudioRef() api.QuestionAudioRef {
	return api.QuestionAudioRef{Url: "", DurationSec: 0}
}

func (b *AudioURLBuilder) urlFor(path string) string {
	if path == "" {
		return ""
	}
	return b.baseURL + "/" + strings.TrimLeft(path, "/")
}
