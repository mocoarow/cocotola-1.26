// Package reference provides service-layer input/output types for workbook reference operations.
package reference

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// ShareWorkbookInput is the validated input for sharing (importing) a workbook.
type ShareWorkbookInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	WorkbookID     string `validate:"required"`
}

// NewShareWorkbookInput creates a validated ShareWorkbookInput.
func NewShareWorkbookInput(operatorID int, organizationID int, workbookID string) (*ShareWorkbookInput, error) {
	m := &ShareWorkbookInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate share workbook input: %w", err)
	}
	return m, nil
}

// ShareWorkbookOutput is the output for a shared workbook reference.
type ShareWorkbookOutput struct {
	ReferenceID string
	WorkbookID  string
	AddedAt     time.Time
}

// ListSharedInput is the validated input for listing shared workbooks.
type ListSharedInput struct {
	OperatorID     int `validate:"required,gt=0"`
	OrganizationID int `validate:"required,gt=0"`
}

// NewListSharedInput creates a validated ListSharedInput.
func NewListSharedInput(operatorID int, organizationID int) (*ListSharedInput, error) {
	m := &ListSharedInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list shared input: %w", err)
	}
	return m, nil
}

// SharedItem represents a shared workbook reference.
type SharedItem struct {
	ReferenceID string
	WorkbookID  string
	AddedAt     time.Time
}

// ListSharedOutput is the output for listing shared workbooks.
type ListSharedOutput struct {
	References []SharedItem
}

// UnshareInput is the validated input for unsharing a workbook.
type UnshareInput struct {
	OperatorID     int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	ReferenceID    string `validate:"required"`
}

// NewUnshareInput creates a validated UnshareInput.
func NewUnshareInput(operatorID int, organizationID int, referenceID string) (*UnshareInput, error) {
	m := &UnshareInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		ReferenceID:    referenceID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate unshare input: %w", err)
	}
	return m, nil
}

// ListPublicInput is the validated input for listing public workbooks.
type ListPublicInput struct {
	OperatorID     int `validate:"required,gt=0"`
	OrganizationID int `validate:"required,gt=0"`
}

// NewListPublicInput creates a validated ListPublicInput.
func NewListPublicInput(operatorID int, organizationID int) (*ListPublicInput, error) {
	m := &ListPublicInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list public input: %w", err)
	}
	return m, nil
}

// PublicWorkbookItem represents a public workbook.
type PublicWorkbookItem struct {
	WorkbookID  string
	OwnerID     int
	Title       string
	Description string
	CreatedAt   time.Time
}

// ListPublicOutput is the output for listing public workbooks.
type ListPublicOutput struct {
	Workbooks []PublicWorkbookItem
}
