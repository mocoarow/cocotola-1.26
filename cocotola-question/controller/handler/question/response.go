package question

import (
	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// ResponseBuilder turns service-layer Item DTOs into wire-level
// api.QuestionResponse values.
type ResponseBuilder struct {
	audio *handler.AudioURLBuilder
}

// NewResponseBuilder returns a builder that renders audio refs with the given builder.
func NewResponseBuilder(audio *handler.AudioURLBuilder) *ResponseBuilder {
	return &ResponseBuilder{audio: audio}
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
		Audio:        b.audio.QuestionAudio(item.Audio),
	}
}
