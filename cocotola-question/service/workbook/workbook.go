// Package workbook provides service-layer input/output types for workbook operations.
package workbook

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// CreateWorkbookInput is the validated input for creating a workbook.
type CreateWorkbookInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	SpaceID        string `validate:"required"`
	Title          string `validate:"required,max=200"`
	Description    string `validate:"max=1000"`
	Visibility     string `validate:"required,oneof=private public"`
}

// NewCreateWorkbookInput creates a validated CreateWorkbookInput.
func NewCreateWorkbookInput(operatorID string, organizationID string, spaceID string, title string, description string, visibility string) (*CreateWorkbookInput, error) {
	m := &CreateWorkbookInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		SpaceID:        spaceID,
		Title:          title,
		Description:    description,
		Visibility:     visibility,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create workbook input: %w", err)
	}
	return m, nil
}

// CreateWorkbookOutput is the output for a created workbook.
type CreateWorkbookOutput struct {
	WorkbookID     string `validate:"required"`
	SpaceID        string `validate:"required"`
	OwnerID        string `validate:"required"`
	OrganizationID string `validate:"required"`
	Title          string `validate:"required"`
	Description    string
	Visibility     string `validate:"required"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewCreateWorkbookOutput creates a validated CreateWorkbookOutput.
func NewCreateWorkbookOutput(workbookID string, spaceID string, ownerID string, organizationID string, title string, description string, visibility string, createdAt time.Time, updatedAt time.Time) (*CreateWorkbookOutput, error) {
	m := &CreateWorkbookOutput{
		WorkbookID:     workbookID,
		SpaceID:        spaceID,
		OwnerID:        ownerID,
		OrganizationID: organizationID,
		Title:          title,
		Description:    description,
		Visibility:     visibility,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create workbook output: %w", err)
	}
	return m, nil
}

// GetWorkbookInput is the validated input for getting a workbook.
type GetWorkbookInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// NewGetWorkbookInput creates a validated GetWorkbookInput.
func NewGetWorkbookInput(operatorID string, organizationID string, workbookID string) (*GetWorkbookInput, error) {
	m := &GetWorkbookInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get workbook input: %w", err)
	}
	return m, nil
}

// Item represents a single workbook in list output.
type Item struct {
	WorkbookID     string
	SpaceID        string
	OwnerID        string
	OrganizationID string
	Title          string
	Description    string
	Visibility     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// GetWorkbookOutput is the output for a single workbook.
type GetWorkbookOutput struct {
	Item
}

// ListWorkbooksInput is the validated input for listing workbooks.
type ListWorkbooksInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	SpaceID        string `validate:"required"`
}

// NewListWorkbooksInput creates a validated ListWorkbooksInput.
func NewListWorkbooksInput(operatorID string, organizationID string, spaceID string) (*ListWorkbooksInput, error) {
	m := &ListWorkbooksInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		SpaceID:        spaceID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list workbooks input: %w", err)
	}
	return m, nil
}

// ListWorkbooksOutput is the output for listing workbooks.
type ListWorkbooksOutput struct {
	Workbooks []Item
}

// UpdateWorkbookInput is the validated input for updating a workbook.
type UpdateWorkbookInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	Title          string `validate:"required,max=200"`
	Description    string `validate:"max=1000"`
	Visibility     string `validate:"required,oneof=private public"`
}

// NewUpdateWorkbookInput creates a validated UpdateWorkbookInput.
func NewUpdateWorkbookInput(operatorID string, organizationID string, workbookID string, title string, description string, visibility string) (*UpdateWorkbookInput, error) {
	m := &UpdateWorkbookInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		Title:          title,
		Description:    description,
		Visibility:     visibility,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate update workbook input: %w", err)
	}
	return m, nil
}

// UpdateWorkbookOutput is the output for an updated workbook.
type UpdateWorkbookOutput struct {
	Item
}

// DeleteWorkbookInput is the validated input for deleting a workbook.
type DeleteWorkbookInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// NewDeleteWorkbookInput creates a validated DeleteWorkbookInput.
func NewDeleteWorkbookInput(operatorID string, organizationID string, workbookID string) (*DeleteWorkbookInput, error) {
	m := &DeleteWorkbookInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate delete workbook input: %w", err)
	}
	return m, nil
}
