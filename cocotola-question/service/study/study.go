// Package study provides service-layer input/output types for study operations.
package study

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// GetStudyQuestionsInput is the validated input for getting study questions.
type GetStudyQuestionsInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	Limit          int    `validate:"gte=1,lte=100"`
}

// NewGetStudyQuestionsInput creates a validated GetStudyQuestionsInput.
func NewGetStudyQuestionsInput(operatorID string, organizationID string, workbookID string, limit int) (*GetStudyQuestionsInput, error) {
	m := &GetStudyQuestionsInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		Limit:          limit,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get study questions input: %w", err)
	}
	return m, nil
}

// QuestionItem represents a question returned for study.
type QuestionItem struct {
	QuestionID   string
	QuestionType string
	Content      string
	Tags         []string
	OrderIndex   int
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
