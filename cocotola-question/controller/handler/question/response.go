package question

import (
	"strings"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// ResponseBuilder turns service-layer Item DTOs into wire-level
// api.QuestionResponse values. It owns presentation concerns such as
// resolving audio paths to public URLs.
type ResponseBuilder struct {
	audioPublicBaseURL string
}

// NewResponseBuilder returns a builder that joins audio paths with the given
// public base URL.
func NewResponseBuilder(audioPublicBaseURL string) *ResponseBuilder {
	return &ResponseBuilder{
		audioPublicBaseURL: strings.TrimRight(audioPublicBaseURL, "/"),
	}
}

// QuestionResponse builds the wire response for a single question item.
func (b *ResponseBuilder) QuestionResponse(item questionservice.Item) api.QuestionResponse {
	return api.QuestionResponse{
		QuestionID:   item.QuestionID,
		QuestionType: item.QuestionType,
		Content:      item.Content,
		Tags:         item.Tags,
		OrderIndex:   int32(item.OrderIndex),
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		Audio:        b.audioFor(item.Audio),
	}
}

func (b *ResponseBuilder) audioFor(audio *questionservice.AudioOutput) api.QuestionAudio {
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

// emptyAudioRef is the placeholder value used when no audio is available for a
// slot. The empty `Url` is the documented "audio not generated" signal in the
// OpenAPI schema — the wire shape always includes the field because codegen
// uses value types (prefer-skip-optional-pointer).
func emptyAudioRef() api.QuestionAudioRef {
	return api.QuestionAudioRef{Url: "", DurationSec: 0}
}

func (b *ResponseBuilder) urlFor(path string) string {
	if path == "" {
		return ""
	}
	return b.audioPublicBaseURL + "/" + strings.TrimLeft(path, "/")
}
