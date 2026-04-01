// Package question provides service-layer input/output types for question operations.
package question

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// AddQuestionInput is the validated input for adding a question.
type AddQuestionInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	WorkbookID     string `validate:"required"`
	QuestionType   string `validate:"required"`
	Content        string `validate:"required,max=10000"`
	OrderIndex     int    `validate:"gte=0"`
}

// NewAddQuestionInput creates a validated AddQuestionInput.
func NewAddQuestionInput(operatorID int, organizationID int, workbookID string, questionType string, content string, orderIndex int) (*AddQuestionInput, error) {
	m := &AddQuestionInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		QuestionType:   questionType,
		Content:        content,
		OrderIndex:     orderIndex,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate add question input: %w", err)
	}
	return m, nil
}

// Item represents a single question.
type Item struct {
	QuestionID   string
	QuestionType string
	Content      string
	OrderIndex   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AddQuestionOutput is the output for an added question.
type AddQuestionOutput struct {
	Item
}

// GetQuestionInput is the validated input for getting a question.
type GetQuestionInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	WorkbookID     string `validate:"required"`
	QuestionID     string `validate:"required"`
}

// NewGetQuestionInput creates a validated GetQuestionInput.
func NewGetQuestionInput(operatorID int, organizationID int, workbookID string, questionID string) (*GetQuestionInput, error) {
	m := &GetQuestionInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		QuestionID:     questionID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get question input: %w", err)
	}
	return m, nil
}

// GetQuestionOutput is the output for a single question.
type GetQuestionOutput struct {
	Item
}

// ListQuestionsInput is the validated input for listing questions.
type ListQuestionsInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	WorkbookID     string `validate:"required"`
}

// NewListQuestionsInput creates a validated ListQuestionsInput.
func NewListQuestionsInput(operatorID int, organizationID int, workbookID string) (*ListQuestionsInput, error) {
	m := &ListQuestionsInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list questions input: %w", err)
	}
	return m, nil
}

// ListQuestionsOutput is the output for listing questions.
type ListQuestionsOutput struct {
	Questions []Item
}

// UpdateQuestionInput is the validated input for updating a question.
type UpdateQuestionInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	WorkbookID     string `validate:"required"`
	QuestionID     string `validate:"required"`
	Content        string `validate:"required,max=10000"`
	OrderIndex     int    `validate:"gte=0"`
}

// NewUpdateQuestionInput creates a validated UpdateQuestionInput.
func NewUpdateQuestionInput(operatorID int, organizationID int, workbookID string, questionID string, content string, orderIndex int) (*UpdateQuestionInput, error) {
	m := &UpdateQuestionInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		QuestionID:     questionID,
		Content:        content,
		OrderIndex:     orderIndex,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate update question input: %w", err)
	}
	return m, nil
}

// UpdateQuestionOutput is the output for an updated question.
type UpdateQuestionOutput struct {
	Item
}

// DeleteQuestionInput is the validated input for deleting a question.
type DeleteQuestionInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	WorkbookID     string `validate:"required"`
	QuestionID     string `validate:"required"`
}

// NewDeleteQuestionInput creates a validated DeleteQuestionInput.
func NewDeleteQuestionInput(operatorID int, organizationID int, workbookID string, questionID string) (*DeleteQuestionInput, error) {
	m := &DeleteQuestionInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		QuestionID:     questionID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate delete question input: %w", err)
	}
	return m, nil
}
