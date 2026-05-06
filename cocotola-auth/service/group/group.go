// Package group provides service-layer types for group management input/output validation.
package group

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	libdomain "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
)

// --- CreateGroup ---

// CreateGroupInput holds the parameters for creating a group.
type CreateGroupInput struct {
	OperatorID       domain.AppUserID
	OrganizationName string `validate:"required"`
	GroupName        string `validate:"required,max=255"`
}

// NewCreateGroupInput creates a validated CreateGroupInput.
func NewCreateGroupInput(operatorID domain.AppUserID, organizationName string, groupName string) (*CreateGroupInput, error) {
	if operatorID.IsZero() {
		return nil, errors.New("create group input operator id must not be zero")
	}
	m := &CreateGroupInput{
		OperatorID:       operatorID,
		OrganizationName: organizationName,
		GroupName:        groupName,
	}
	if err := libdomain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create group input: %w", err)
	}
	return m, nil
}

// CreateGroupOutput holds the result of creating a group.
type CreateGroupOutput struct {
	GroupID        domain.GroupID
	OrganizationID domain.OrganizationID
	Name           string `validate:"required"`
	Enabled        bool
}

// NewCreateGroupOutput creates a validated CreateGroupOutput.
func NewCreateGroupOutput(groupID domain.GroupID, organizationID domain.OrganizationID, name string, enabled bool) (*CreateGroupOutput, error) {
	if groupID.IsZero() {
		return nil, errors.New("create group output group id must not be zero")
	}
	if organizationID.IsZero() {
		return nil, errors.New("create group output organization id must not be zero")
	}
	m := &CreateGroupOutput{
		GroupID:        groupID,
		OrganizationID: organizationID,
		Name:           name,
		Enabled:        enabled,
	}
	if err := libdomain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create group output: %w", err)
	}
	return m, nil
}
