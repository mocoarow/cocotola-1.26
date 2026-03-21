// Package group provides service-layer types for group management input/output validation.
package group

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// --- CreateGroup ---

// CreateGroupInput holds the parameters for creating a group.
type CreateGroupInput struct {
	OperatorID       int    `validate:"required,gt=0"`
	OrganizationName string `validate:"required"`
	GroupName        string `validate:"required,max=255"`
}

// NewCreateGroupInput creates a validated CreateGroupInput.
func NewCreateGroupInput(operatorID int, organizationName string, groupName string) (*CreateGroupInput, error) {
	m := &CreateGroupInput{
		OperatorID:       operatorID,
		OrganizationName: organizationName,
		GroupName:        groupName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create group input: %w", err)
	}
	return m, nil
}

// CreateGroupOutput holds the result of creating a group.
type CreateGroupOutput struct {
	GroupID        int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	Name           string `validate:"required"`
	Enabled        bool
}

// NewCreateGroupOutput creates a validated CreateGroupOutput.
func NewCreateGroupOutput(groupID int, organizationID int, name string, enabled bool) (*CreateGroupOutput, error) {
	m := &CreateGroupOutput{
		GroupID:        groupID,
		OrganizationID: organizationID,
		Name:           name,
		Enabled:        enabled,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create group output: %w", err)
	}
	return m, nil
}
