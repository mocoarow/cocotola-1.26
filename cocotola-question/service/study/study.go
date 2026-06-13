// Package study provides service-layer input/output types for study operations.
package study

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// MaxExcludeIDsCount caps the size of the client-supplied resume set
// (questions already answered in the in-progress browser session) so a
// malicious or buggy client cannot send an unbounded list.
const MaxExcludeIDsCount = 1000

// MaxExcludeIDLength caps each entry in the resume set. Mirrors
// MaxChoiceIDLength so the two question-ID inputs hit the same upper bound,
// keeping per-request memory bounded even if the count cap is reached.
const MaxExcludeIDLength = 100

// GetStudyQuestionsInput is the validated input for getting study questions.
//
// Practice is an off-the-record mode: when true, the usecase ignores the
// per-question NextDueAt schedule and returns every active question. Callers
// using this mode are expected to skip the "record answer" endpoint so the
// user's spaced-repetition records and counters stay untouched.
//
// ExcludeIDs lets the client resume an interrupted study session by skipping
// questions it has already answered locally. The server is stateless; the
// browser tracks the in-progress set and discards it at the next local 03:00
// boundary.
type GetStudyQuestionsInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	Limit          int    `validate:"gte=1,lte=100"`
	Practice       bool
	ExcludeIDs     []string
}

// NewGetStudyQuestionsInput creates a validated GetStudyQuestionsInput.
func NewGetStudyQuestionsInput(operatorID string, organizationID string, workbookID string, limit int, practice bool, excludeIDs []string) (*GetStudyQuestionsInput, error) {
	if len(excludeIDs) > MaxExcludeIDsCount {
		return nil, fmt.Errorf("excludeIds count exceeds limit (max %d, got %d): %w", MaxExcludeIDsCount, len(excludeIDs), domain.ErrInvalidArgument)
	}
	for i, id := range excludeIDs {
		if id == "" {
			return nil, fmt.Errorf("excludeIds[%d] is empty: %w", i, domain.ErrInvalidArgument)
		}
		if len(id) > MaxExcludeIDLength {
			return nil, fmt.Errorf("excludeIds[%d] exceeds length limit (max %d, got %d): %w", i, MaxExcludeIDLength, len(id), domain.ErrInvalidArgument)
		}
	}
	m := &GetStudyQuestionsInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		Limit:          limit,
		Practice:       practice,
		ExcludeIDs:     excludeIDs,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get study questions input: %w", err)
	}
	return m, nil
}

// GetStudySummaryInput is the validated input for getting study summary counts.
//
// Practice mirrors GetStudyQuestionsInput.Practice: when true the summary is
// computed without applying the per-question NextDueAt schedule, so callers
// rendering a "practice mode" picker see the unrestricted available pool.
type GetStudySummaryInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	Practice       bool
}

// NewGetStudySummaryInput creates a validated GetStudySummaryInput.
func NewGetStudySummaryInput(operatorID, organizationID, workbookID string, practice bool) (*GetStudySummaryInput, error) {
	m := &GetStudySummaryInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		Practice:       practice,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get study summary input: %w", err)
	}
	return m, nil
}

// GetStudySummaryOutput is the output for getting study summary counts. The
// ratio fields advertise the server-side review/new mix so clients can render
// it in the picker without hard-coding the same constants.
type GetStudySummaryOutput struct {
	NewCount               int
	ReviewCount            int
	TotalDue               int
	ReviewRatioNumerator   int
	ReviewRatioDenominator int
}

// QuestionItem represents a question returned for study.
type QuestionItem struct {
	QuestionID   string
	QuestionType string
	Content      string
	Tags         []string
	OrderIndex   int
	// Audio carries playable audio metadata when the audio batch has finished.
	// Nil when no audio is ready (the typical state for a freshly created
	// word_fill question).
	Audio *QuestionItemAudio
}

// QuestionItemAudio mirrors the same shape as questionservice.AudioOutput but
// avoids cross-package coupling: study items are independent of the question
// CRUD response shape.
type QuestionItemAudio struct {
	Source *QuestionItemAudioRef
	Target *QuestionItemAudioRef
}

// QuestionItemAudioRef carries the storage-relative path and metadata of a
// single audio file. The handler turns Path into a public URL.
type QuestionItemAudioRef struct {
	Path        string
	DurationSec float64
}

// GetStudyQuestionsOutput is the output for getting study questions.
type GetStudyQuestionsOutput struct {
	Questions   []QuestionItem
	TotalDue    int
	NewCount    int
	ReviewCount int
}

// Limits on the multiple_choice answer payload. These mirror the OpenAPI
// schema (maxItems / maxLength on selectedChoiceIds) and are enforced here so
// the contract holds even when callers bypass the spec.
const (
	MaxSelectedChoiceIDsCount = 40
	MaxChoiceIDLength         = 100
)

// RecordAnswerInput is the validated input for recording an answer.
// Exactly one of Correct or SelectedChoiceIDs must be set, matched to the
// question's type. The usecase enforces the per-type rule once the question
// is loaded (the handler also rejects "neither set" / "both set" earlier).
type RecordAnswerInput struct {
	OperatorID        string `validate:"required"`
	OrganizationID    string `validate:"required"`
	WorkbookID        string `validate:"required"`
	QuestionID        string `validate:"required"`
	Correct           *bool
	SelectedChoiceIDs *[]string
}

// NewRecordAnswerInputForWordFill creates a validated RecordAnswerInput for word_fill questions.
func NewRecordAnswerInputForWordFill(operatorID, organizationID, workbookID, questionID string, correct bool) (*RecordAnswerInput, error) {
	m := &RecordAnswerInput{
		OperatorID:        operatorID,
		OrganizationID:    organizationID,
		WorkbookID:        workbookID,
		QuestionID:        questionID,
		Correct:           &correct,
		SelectedChoiceIDs: nil,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate record answer input: %w", err)
	}
	return m, nil
}

// NewRecordAnswerInputForMultipleChoice creates a validated RecordAnswerInput for multiple_choice questions.
func NewRecordAnswerInputForMultipleChoice(operatorID, organizationID, workbookID, questionID string, selectedChoiceIDs []string) (*RecordAnswerInput, error) {
	if len(selectedChoiceIDs) > MaxSelectedChoiceIDsCount {
		return nil, fmt.Errorf("selectedChoiceIds count exceeds limit (max %d, got %d): %w", MaxSelectedChoiceIDsCount, len(selectedChoiceIDs), domain.ErrInvalidArgument)
	}
	for i, id := range selectedChoiceIDs {
		if len(id) > MaxChoiceIDLength {
			return nil, fmt.Errorf("selectedChoiceIds[%d] exceeds length limit (max %d, got %d): %w", i, MaxChoiceIDLength, len(id), domain.ErrInvalidArgument)
		}
	}
	ids := selectedChoiceIDs
	m := &RecordAnswerInput{
		OperatorID:        operatorID,
		OrganizationID:    organizationID,
		WorkbookID:        workbookID,
		QuestionID:        questionID,
		Correct:           nil,
		SelectedChoiceIDs: &ids,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate record answer input: %w", err)
	}
	return m, nil
}

// RecordAnswerOutput is the output for recording an answer.
type RecordAnswerOutput struct {
	NextDueAt          time.Time
	ConsecutiveCorrect int
	TotalCorrect       int
	TotalIncorrect     int
}

// DeleteStudyHistoryInput is the validated input for clearing a workbook's
// study history. The operation only affects the operator's own records.
type DeleteStudyHistoryInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// ListStudyRecordsInput is the validated input for listing a workbook's
// study records belonging to the operator.
type ListStudyRecordsInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// NewListStudyRecordsInput creates a validated ListStudyRecordsInput.
func NewListStudyRecordsInput(operatorID, organizationID, workbookID string) (*ListStudyRecordsInput, error) {
	m := &ListStudyRecordsInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list study records input: %w", err)
	}
	return m, nil
}

// RecordItem mirrors a single study record at the service boundary,
// decoupling controllers from the domain aggregate.
type RecordItem struct {
	WorkbookID         string
	QuestionID         string
	ConsecutiveCorrect int
	LastAnsweredAt     time.Time
	NextDueAt          time.Time
	TotalCorrect       int
	TotalIncorrect     int
}

// ListStudyRecordsOutput holds the records returned for a workbook.
type ListStudyRecordsOutput struct {
	Records []RecordItem
}

// NewDeleteStudyHistoryInput creates a validated DeleteStudyHistoryInput.
func NewDeleteStudyHistoryInput(operatorID, organizationID, workbookID string) (*DeleteStudyHistoryInput, error) {
	m := &DeleteStudyHistoryInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate delete study history input: %w", err)
	}
	return m, nil
}
