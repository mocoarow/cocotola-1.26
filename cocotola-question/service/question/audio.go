package question

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// PendingAudioItem describes one queue entry the audio batch should generate.
// Targets that are not word_fill are never returned here.
type PendingAudioItem struct {
	WorkbookID  string
	QuestionID  string
	SourceText  string
	SourceLang  string
	TargetText  string
	TargetLang  string
	InputHash   string
	FailedTries int
	UpdatedAt   time.Time
}

// ListPendingAudioInput holds parameters for paging through pending audio items.
type ListPendingAudioInput struct {
	Limit int `validate:"required,gte=1,lte=200"`
}

// NewListPendingAudioInput creates a validated ListPendingAudioInput.
func NewListPendingAudioInput(limit int) (*ListPendingAudioInput, error) {
	m := &ListPendingAudioInput{Limit: limit}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list pending audio input: %w: %w", domain.ErrInvalidArgument, err)
	}
	return m, nil
}

// ListPendingAudioOutput carries pending queue entries.
type ListPendingAudioOutput struct {
	Items []PendingAudioItem
}

// ClaimAudioInput is the input for transitioning a question's audio status to
// generating, with optimistic protection via the input hash.
//
// InputHash must be a sha256 digest in lowercase hex (exactly 64 hex chars).
type ClaimAudioInput struct {
	WorkbookID string `validate:"required"`
	QuestionID string `validate:"required"`
	InputHash  string `validate:"required,len=64,hexadecimal"`
}

// NewClaimAudioInput creates a validated ClaimAudioInput.
func NewClaimAudioInput(workbookID, questionID, inputHash string) (*ClaimAudioInput, error) {
	m := &ClaimAudioInput{
		WorkbookID: workbookID,
		QuestionID: questionID,
		InputHash:  inputHash,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate claim audio input: %w: %w", domain.ErrInvalidArgument, err)
	}
	return m, nil
}

// CompleteAudioRefInput describes a single generated audio file.
type CompleteAudioRefInput struct {
	Path        string  `validate:"required"`
	DurationSec float64 `validate:"gte=0"`
	SizeBytes   int64   `validate:"gte=0"`
}

// CompleteAudioInput is the input for transitioning a question's audio status
// from generating to ready.
//
// InputHash must be a sha256 digest in lowercase hex (exactly 64 hex chars).
type CompleteAudioInput struct {
	WorkbookID string                           `validate:"required"`
	QuestionID string                           `validate:"required"`
	InputHash  string                           `validate:"required,len=64,hexadecimal"`
	Refs       map[string]CompleteAudioRefInput `validate:"required,dive"`
}

// NewCompleteAudioInput creates a validated CompleteAudioInput.
func NewCompleteAudioInput(workbookID, questionID, inputHash string, refs map[string]CompleteAudioRefInput) (*CompleteAudioInput, error) {
	m := &CompleteAudioInput{
		WorkbookID: workbookID,
		QuestionID: questionID,
		InputHash:  inputHash,
		Refs:       refs,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate complete audio input: %w: %w", domain.ErrInvalidArgument, err)
	}
	return m, nil
}

// FailAudioInput is the input for transitioning a question's audio status to
// failed (retried by a future batch run).
//
// InputHash must be a sha256 digest in lowercase hex (exactly 64 hex chars).
type FailAudioInput struct {
	WorkbookID string `validate:"required"`
	QuestionID string `validate:"required"`
	InputHash  string `validate:"required,len=64,hexadecimal"`
	Reason     string `validate:"max=500"`
}

// NewFailAudioInput creates a validated FailAudioInput.
func NewFailAudioInput(workbookID, questionID, inputHash, reason string) (*FailAudioInput, error) {
	m := &FailAudioInput{
		WorkbookID: workbookID,
		QuestionID: questionID,
		InputHash:  inputHash,
		Reason:     reason,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate fail audio input: %w: %w", domain.ErrInvalidArgument, err)
	}
	return m, nil
}

// ReclaimStaleAudioInput is the input for the reaper that returns "generating"
// items stuck longer than StaleAfter back to "pending" so they get retried.
type ReclaimStaleAudioInput struct {
	StaleAfter time.Duration `validate:"required"`
	Limit      int           `validate:"required,gte=1,lte=200"`
}

// NewReclaimStaleAudioInput creates a validated ReclaimStaleAudioInput.
func NewReclaimStaleAudioInput(staleAfter time.Duration, limit int) (*ReclaimStaleAudioInput, error) {
	m := &ReclaimStaleAudioInput{
		StaleAfter: staleAfter,
		Limit:      limit,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate reclaim stale audio input: %w: %w", domain.ErrInvalidArgument, err)
	}
	return m, nil
}

// ReclaimStaleAudioOutput reports how many entries were reclaimed.
type ReclaimStaleAudioOutput struct {
	Reclaimed int
}
