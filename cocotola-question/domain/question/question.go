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

// Question is the aggregate root persisted by QuestionRepository.
// It references the parent Workbook by workbookID.
type Question struct {
	id           string
	workbookID   string
	questionType Type
	content      string
	tags         []string
	orderIndex   int
	version      int
	createdAt    time.Time
	updatedAt    time.Time
}

// NewQuestion creates a validated Question with version=0 (a new aggregate not yet saved).
// Callers (usecase layer) must supply the ID and timestamps.
func NewQuestion(id string, workbookID string, questionType Type, content string, tags []string, orderIndex int, createdAt time.Time, updatedAt time.Time) (*Question, error) {
	q := &Question{
		id:           id,
		workbookID:   workbookID,
		questionType: questionType,
		content:      content,
		tags:         copyTags(tags),
		orderIndex:   orderIndex,
		version:      0,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
	if err := q.validate(); err != nil {
		return nil, fmt.Errorf("new question: %w", err)
	}
	return q, nil
}

// ReconstructQuestion reconstitutes a Question from persistence without validation.
func ReconstructQuestion(id string, workbookID string, questionType Type, content string, tags []string, orderIndex int, version int, createdAt time.Time, updatedAt time.Time) *Question {
	return &Question{
		id:           id,
		workbookID:   workbookID,
		questionType: questionType,
		content:      content,
		tags:         copyTags(tags),
		orderIndex:   orderIndex,
		version:      version,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Edit updates content, tags, and orderIndex with validation.
// Callers (usecase layer) must supply the new updatedAt timestamp.
func (q *Question) Edit(content string, tags []string, orderIndex int, updatedAt time.Time) error {
	cp := copyTags(tags)
	candidate := &Question{
		id:           q.id,
		workbookID:   q.workbookID,
		questionType: q.questionType,
		content:      content,
		tags:         cp,
		orderIndex:   orderIndex,
		version:      q.version,
		createdAt:    q.createdAt,
		updatedAt:    updatedAt,
	}
	if err := candidate.validate(); err != nil {
		return fmt.Errorf("edit question: %w", err)
	}
	q.content = content
	q.tags = cp
	q.orderIndex = orderIndex
	q.updatedAt = updatedAt
	return nil
}

func (q *Question) validate() error {
	if q.id == "" {
		return fmt.Errorf("question id is required: %w", domain.ErrInvalidArgument)
	}
	if q.workbookID == "" {
		return fmt.Errorf("question workbook id is required: %w", domain.ErrInvalidArgument)
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
		return fmt.Errorf("validate tags: %w", err)
	}
	if err := ValidateContent(q.questionType, q.content); err != nil {
		return fmt.Errorf("validate content: %w", err)
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

// WorkbookID returns the parent workbook ID.
func (q *Question) WorkbookID() string { return q.workbookID }

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

// Version returns the persisted version (0 = new, not yet saved).
func (q *Question) Version() int { return q.version }

// SetVersion sets the persisted version on the aggregate.
// Intended for repository implementations to update the version after a successful save.
// Do not call from application or domain code.
func (q *Question) SetVersion(version int) { q.version = version }

// CreatedAt returns the creation timestamp.
func (q *Question) CreatedAt() time.Time { return q.createdAt }

// UpdatedAt returns the last update timestamp.
func (q *Question) UpdatedAt() time.Time { return q.updatedAt }
