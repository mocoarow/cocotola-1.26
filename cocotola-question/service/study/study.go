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

// RecordAnswerInput is the validated input for recording an answer.
type RecordAnswerInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	QuestionID     string `validate:"required"`
	Correct        bool
}

// NewRecordAnswerInput creates a validated RecordAnswerInput.
func NewRecordAnswerInput(operatorID string, organizationID string, workbookID string, questionID string, correct bool) (*RecordAnswerInput, error) {
	m := &RecordAnswerInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		QuestionID:     questionID,
		Correct:        correct,
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
