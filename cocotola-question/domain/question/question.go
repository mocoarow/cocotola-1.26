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
type Question struct {
	id              string
	workbookID      string
	questionType    Type
	content         string
	tags            []string
	orderIndex      int
	version         int
	createdAt       time.Time
	updatedAt       time.Time
	audioGeneration *AudioGeneration
	// nil for non-word_fill types or aggregates from ReconstructQuestion.
	parsedWordFill *WordFillContent
}

// NewQuestion creates a validated Question with version=0 (a new aggregate not yet saved).
// Callers (usecase layer) must supply the ID and timestamps.
func NewQuestion(parser ContentParser, id string, workbookID string, questionType Type, content string, tags []string, orderIndex int, createdAt time.Time, updatedAt time.Time) (*Question, error) {
	q := &Question{
		id:              id,
		workbookID:      workbookID,
		questionType:    questionType,
		content:         content,
		tags:            copyTags(tags),
		orderIndex:      orderIndex,
		version:         0,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		audioGeneration: nil,
		parsedWordFill:  nil,
	}
	if err := q.validate(parser); err != nil {
		return nil, fmt.Errorf("new question: %w", err)
	}
	return q, nil
}

// ReconstructQuestion reconstitutes a Question from persistence without validation.
func ReconstructQuestion(id string, workbookID string, questionType Type, content string, tags []string, orderIndex int, version int, createdAt time.Time, updatedAt time.Time) *Question {
	return &Question{
		id:              id,
		workbookID:      workbookID,
		questionType:    questionType,
		content:         content,
		tags:            copyTags(tags),
		orderIndex:      orderIndex,
		version:         version,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		audioGeneration: nil,
		parsedWordFill:  nil,
	}
}

// Edit updates content, tags, and orderIndex with validation.
// Callers (usecase layer) must supply the new updatedAt timestamp.
func (q *Question) Edit(parser ContentParser, content string, tags []string, orderIndex int, updatedAt time.Time) error {
	cp := copyTags(tags)
	candidate := &Question{
		id:              q.id,
		workbookID:      q.workbookID,
		questionType:    q.questionType,
		content:         content,
		tags:            cp,
		orderIndex:      orderIndex,
		version:         q.version,
		createdAt:       q.createdAt,
		updatedAt:       updatedAt,
		audioGeneration: q.audioGeneration,
		parsedWordFill:  nil,
	}
	if err := candidate.validate(parser); err != nil {
		return fmt.Errorf("edit question: %w", err)
	}
	q.content = content
	q.tags = cp
	q.orderIndex = orderIndex
	q.updatedAt = updatedAt
	q.parsedWordFill = candidate.parsedWordFill
	return nil
}

func (q *Question) validate(parser ContentParser) error {
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
	return q.validateContent(parser)
}

func (q *Question) validateContent(parser ContentParser) error {
	switch q.questionType.Value() {
	case questionTypeWordFill:
		c, err := parser.ParseWordFill(q.content)
		if err != nil {
			return fmt.Errorf("parse word_fill content: %w", err)
		}
		if err := c.Validate(); err != nil {
			return fmt.Errorf("validate word_fill content: %w", err)
		}
		q.parsedWordFill = c
		return nil
	case questionTypeMultipleChoice:
		c, err := parser.ParseMultipleChoice(q.content)
		if err != nil {
			return fmt.Errorf("parse multiple_choice content: %w", err)
		}
		if err := c.Validate(); err != nil {
			return fmt.Errorf("validate multiple_choice content: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown question type for content validation: %w", domain.ErrInvalidArgument)
	}
}

// WordFillContentParsed returns the cached parsed word_fill content populated
// during validation. Returns nil for non-word_fill questions or for aggregates
// produced via ReconstructQuestion (which skips validation and the cache).
func (q *Question) WordFillContentParsed() *WordFillContent { return q.parsedWordFill }

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

// AudioGeneration returns the shared audio-generation state (nil when unset); do not mutate it.
func (q *Question) AudioGeneration() *AudioGeneration { return q.audioGeneration }

// SetAudioGeneration replaces the audio-generation state.
// Intended for repository hydration and audio-batch transitions only; the
// question-edit usecase should use MarkAudioPending instead.
func (q *Question) SetAudioGeneration(ag *AudioGeneration) { q.audioGeneration = ag }

// MarkAudioPending sets the audio-generation state to pending; a no-op for the
// same inputHash so re-saves do not bump the queue position.
func (q *Question) MarkAudioPending(inputHash string, now time.Time) {
	if q.audioGeneration != nil && q.audioGeneration.inputHash == inputHash {
		return
	}
	q.audioGeneration = &AudioGeneration{
		status:         AudioGenerationStatusPending(),
		inputHash:      inputHash,
		refs:           nil,
		updatedAt:      now,
		failedAttempts: 0,
		lastError:      "",
	}
}

// ClaimAudio transitions the audio generation from pending to generating.
// Race protection comes from the optimistic lock in Save, not from this check.
func (q *Question) ClaimAudio(inputHash string, now time.Time) error {
	if q.audioGeneration == nil || !q.audioGeneration.status.IsPending() {
		return domain.ErrAudioNotPending
	}
	if q.audioGeneration.inputHash != inputHash {
		return domain.ErrAudioInputHashMismatch
	}
	q.audioGeneration = &AudioGeneration{
		status:         AudioGenerationStatusGenerating(),
		inputHash:      q.audioGeneration.inputHash,
		refs:           q.audioGeneration.Refs(),
		updatedAt:      now,
		failedAttempts: q.audioGeneration.failedAttempts,
		lastError:      "",
	}
	return nil
}

// CompleteAudio transitions the audio generation from generating to ready.
func (q *Question) CompleteAudio(inputHash string, refs map[string]AudioRef, now time.Time) error {
	if q.audioGeneration == nil || q.audioGeneration.status.Value() != AudioGenerationStatusGenerating().Value() {
		return domain.ErrAudioNotGenerating
	}
	if q.audioGeneration.inputHash != inputHash {
		return domain.ErrAudioInputHashMismatch
	}
	q.audioGeneration = &AudioGeneration{
		status:         AudioGenerationStatusReady(),
		inputHash:      inputHash,
		refs:           copyAudioRefs(refs),
		updatedAt:      now,
		failedAttempts: 0,
		lastError:      "",
	}
	return nil
}

// FailAudio transitions the audio generation from generating to failed and
// increments the failure counter used for backoff.
func (q *Question) FailAudio(inputHash string, reason string, now time.Time) error {
	if q.audioGeneration == nil || q.audioGeneration.status.Value() != AudioGenerationStatusGenerating().Value() {
		return domain.ErrAudioNotGenerating
	}
	if q.audioGeneration.inputHash != inputHash {
		return domain.ErrAudioInputHashMismatch
	}
	q.audioGeneration = &AudioGeneration{
		status:         AudioGenerationStatusFailed(),
		inputHash:      q.audioGeneration.inputHash,
		refs:           q.audioGeneration.Refs(),
		updatedAt:      now,
		failedAttempts: q.audioGeneration.failedAttempts + 1,
		lastError:      truncateRunes(reason, MaxAudioLastErrorLength),
	}
	return nil
}

// ReclaimStaleAudio moves a "generating" entry older than staleAfter back to
// "pending" so a crashed worker does not freeze it; true means persist.
func (q *Question) ReclaimStaleAudio(now time.Time, staleAfter time.Duration) bool {
	if q.audioGeneration == nil {
		return false
	}
	if q.audioGeneration.status.Value() != AudioGenerationStatusGenerating().Value() {
		return false
	}
	if now.Sub(q.audioGeneration.updatedAt) < staleAfter {
		return false
	}
	q.audioGeneration = &AudioGeneration{
		status:         AudioGenerationStatusPending(),
		inputHash:      q.audioGeneration.inputHash,
		refs:           q.audioGeneration.Refs(),
		updatedAt:      now,
		failedAttempts: q.audioGeneration.failedAttempts,
		lastError:      q.audioGeneration.lastError,
	}
	return true
}
