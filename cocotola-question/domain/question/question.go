package question

import (
	"errors"
	"fmt"
	"time"
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
		return errors.New("question id is required")
	}
	if q.questionType.Value() == "" {
		return errors.New("question type is required")
	}
	if q.content == "" {
		return errors.New("question content is required")
	}
	if len(q.content) > maxContentLength {
		return fmt.Errorf("question content must not exceed %d characters", maxContentLength)
	}
	if q.orderIndex < 0 {
		return errors.New("question order index must not be negative")
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
