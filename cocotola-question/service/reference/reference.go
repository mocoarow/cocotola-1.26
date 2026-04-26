// Package reference provides service-layer input/output types for workbook reference operations.
package reference

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// ShareWorkbookInput is the validated input for sharing (importing) a workbook.
type ShareWorkbookInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// NewShareWorkbookInput creates a validated ShareWorkbookInput.
func NewShareWorkbookInput(operatorID string, organizationID string, workbookID string) (*ShareWorkbookInput, error) {
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
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
}

// NewListSharedInput creates a validated ListSharedInput.
func NewListSharedInput(operatorID string, organizationID string) (*ListSharedInput, error) {
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
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	ReferenceID    string `validate:"required"`
}

// NewUnshareInput creates a validated UnshareInput.
func NewUnshareInput(operatorID string, organizationID string, referenceID string) (*UnshareInput, error) {
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
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	Language       string `validate:"required,len=2"`
}

// NewListPublicInput creates a validated ListPublicInput.
func NewListPublicInput(operatorID string, organizationID string, language string) (*ListPublicInput, error) {
	m := &ListPublicInput{
		OperatorID:     operatorID,
		OrganizationID: organizationID,
		Language:       language,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list public input: %w", err)
	}
	return m, nil
}

// PublicWorkbookItem represents a public workbook.
type PublicWorkbookItem struct {
	WorkbookID  string
	OwnerID     string
	Title       string
	Description string
	Language    string
	CreatedAt   time.Time
}

// ListPublicOutput is the output for listing public workbooks.
type ListPublicOutput struct {
	Workbooks []PublicWorkbookItem
}
