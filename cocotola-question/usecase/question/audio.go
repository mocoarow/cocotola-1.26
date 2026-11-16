package question

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// markAudioPendingIfWordFill queues word_fill audio generation; safe on every
// save because MarkAudioPending ignores an unchanged inputHash.
func markAudioPendingIfWordFill(q *domainquestion.Question, now time.Time) error {
	if q.QuestionType().Value() != domainquestion.TypeWordFill().Value() {
		return nil
	}
	c := q.WordFillContentParsed()
	if c == nil {
		parsed, err := contentjson.Parser{}.ParseWordFill(q.Content())
		if err != nil {
			return fmt.Errorf("parse word_fill content: %w", err)
		}
		c = parsed
	}
	hash := domainquestion.ComputeWordFillAudioInputHash(*c)
	q.MarkAudioPending(hash, now)
	return nil
}

// toServiceItem projects a domain Question into the service-layer Item DTO,
// including audio refs when ready.
func toServiceItem(q *domainquestion.Question) questionservice.Item {
	return questionservice.Item{
		QuestionID:   q.ID(),
		QuestionType: q.QuestionType().Value(),
		Content:      q.Content(),
		Tags:         q.Tags(),
		OrderIndex:   q.OrderIndex(),
		CreatedAt:    q.CreatedAt(),
		UpdatedAt:    q.UpdatedAt(),
		Audio:        questionservice.AudioOutputFromRefs(q.ReadyAudioRefs()),
	}
}
