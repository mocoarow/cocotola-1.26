package question

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	maxContentLength = 10000
)

// Question is an entity within the Workbook aggregate.
type Question struct {
	id           string
	questionType Type
	content      string
	orderIndex   int
	createdAt    time.Time
	updatedAt    time.Time
}

// NewQuestion creates a validated Question.
func NewQuestion(id string, questionType Type, content string, orderIndex int, createdAt time.Time, updatedAt time.Time) (*Question, error) {
	m := &Question{
		id:           id,
		questionType: questionType,
		content:      content,
		orderIndex:   orderIndex,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructQuestion reconstitutes a Question from persistence without validation.
func ReconstructQuestion(id string, questionType Type, content string, orderIndex int, createdAt time.Time, updatedAt time.Time) *Question {
	return &Question{
		id:           id,
		questionType: questionType,
		content:      content,
		orderIndex:   orderIndex,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (q *Question) validate() error {
	if q.id == "" {
		return fmt.Errorf("question id is required: %w", domain.ErrInvalidArgument)
	}
	if q.questionType.Value() == "" {
		return fmt.Errorf("question type is required: %w", domain.ErrInvalidArgument)
	}
	if q.content == "" {
		return fmt.Errorf("question content is required: %w", domain.ErrInvalidArgument)
	}
	if len(q.content) > maxContentLength {
		return fmt.Errorf("question content must not exceed %d characters: %w", maxContentLength, domain.ErrInvalidArgument)
	}
	if q.orderIndex < 0 {
		return fmt.Errorf("question order index must not be negative: %w", domain.ErrInvalidArgument)
	}
	return nil
}

// ID returns the question ID.
func (q *Question) ID() string { return q.id }

// QuestionType returns the question type.
func (q *Question) QuestionType() Type { return q.questionType }

// Content returns the question content.
func (q *Question) Content() string { return q.content }

// OrderIndex returns the display order.
func (q *Question) OrderIndex() int { return q.orderIndex }

// CreatedAt returns the creation timestamp.
func (q *Question) CreatedAt() time.Time { return q.createdAt }

// UpdatedAt returns the last update timestamp.
func (q *Question) UpdatedAt() time.Time { return q.updatedAt }
