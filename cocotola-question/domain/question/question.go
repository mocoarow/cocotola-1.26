package question

import (
	"fmt"
	"regexp"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	maxContentLength = 10000
	maxTags          = 20
	maxTagLength     = 100
)

var tagPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+$`)

// Question is an entity within the Workbook aggregate.
type Question struct {
	id           string
	questionType Type
	content      string
	tags         []string
	orderIndex   int
	createdAt    time.Time
	updatedAt    time.Time
}

// NewQuestion creates a validated Question.
func NewQuestion(id string, questionType Type, content string, tags []string, orderIndex int, createdAt time.Time, updatedAt time.Time) (*Question, error) {
	m := &Question{
		id:           id,
		questionType: questionType,
		content:      content,
		tags:         copyTags(tags),
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
func ReconstructQuestion(id string, questionType Type, content string, tags []string, orderIndex int, createdAt time.Time, updatedAt time.Time) *Question {
	return &Question{
		id:           id,
		questionType: questionType,
		content:      content,
		tags:         copyTags(tags),
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
	if err := validateTags(q.tags); err != nil {
		return err
	}
	if err := ValidateContent(q.questionType, q.content); err != nil {
		return err
	}
	return nil
}

func validateTags(tags []string) error {
	if len(tags) > maxTags {
		return fmt.Errorf("tags must not exceed %d: %w", maxTags, domain.ErrInvalidArgument)
	}
	seen := make(map[string]bool, len(tags))
	for _, tag := range tags {
		if len(tag) > maxTagLength {
			return fmt.Errorf("tag must not exceed %d characters: %w", maxTagLength, domain.ErrInvalidArgument)
		}
		if !tagPattern.MatchString(tag) {
			return fmt.Errorf("tag %q must match format 'key:value': %w", tag, domain.ErrInvalidArgument)
		}
		if seen[tag] {
			return fmt.Errorf("duplicate tag %q: %w", tag, domain.ErrInvalidArgument)
		}
		seen[tag] = true
	}
	return nil
}

// ID returns the question ID.
func (q *Question) ID() string { return q.id }

// QuestionType returns the question type.
func (q *Question) QuestionType() Type { return q.questionType }

// Content returns the question content.
func (q *Question) Content() string { return q.content }

// Tags returns the question tags.
func (q *Question) Tags() []string { return copyTags(q.tags) }

func copyTags(tags []string) []string {
	if tags == nil {
		return nil
	}
	cp := make([]string, len(tags))
	copy(cp, tags)
	return cp
}

// OrderIndex returns the display order.
func (q *Question) OrderIndex() int { return q.orderIndex }

// CreatedAt returns the creation timestamp.
func (q *Question) CreatedAt() time.Time { return q.createdAt }

// UpdatedAt returns the last update timestamp.
func (q *Question) UpdatedAt() time.Time { return q.updatedAt }
