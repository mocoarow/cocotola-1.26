// Package space provides service-layer types for space management input/output validation.
package space

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// --- CreateSpace ---

// CreateSpaceInput holds the parameters for creating a space.
type CreateSpaceInput struct {
	OperatorID       domain.AppUserID
	OrganizationName string `validate:"required"`
	Name             string `validate:"required,max=100"`
	SpaceType        string `validate:"required,oneof=public private"`
}

// NewCreateSpaceInput creates a validated CreateSpaceInput.
func NewCreateSpaceInput(operatorID domain.AppUserID, organizationName string, name string, spaceType string) (*CreateSpaceInput, error) {
	if operatorID.IsZero() {
		return nil, errors.New("create space input operator id must not be zero")
	}
	m := &CreateSpaceInput{
		OperatorID:       operatorID,
		OrganizationName: organizationName,
		Name:             name,
		SpaceType:        spaceType,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create space input: %w", err)
	}
	return m, nil
}

// CreateSpaceOutput holds the result of creating a space.
// SpaceID remains int in Phase 1.
type CreateSpaceOutput struct {
	SpaceID        int `validate:"required,gt=0"`
	OrganizationID domain.OrganizationID
	OwnerID        domain.AppUserID
	KeyName        string `validate:"required"`
	Name           string `validate:"required"`
	SpaceType      string `validate:"required"`
	Deleted        bool
}

// NewCreateSpaceOutput creates a validated CreateSpaceOutput.
func NewCreateSpaceOutput(spaceID int, organizationID domain.OrganizationID, ownerID domain.AppUserID, keyName string, name string, spaceType string, deleted bool) (*CreateSpaceOutput, error) {
	if organizationID.IsZero() {
		return nil, errors.New("create space output organization id must not be zero")
	}
	if ownerID.IsZero() {
		return nil, errors.New("create space output owner id must not be zero")
	}
	m := &CreateSpaceOutput{
		SpaceID:        spaceID,
		OrganizationID: organizationID,
		OwnerID:        ownerID,
		KeyName:        keyName,
		Name:           name,
		SpaceType:      spaceType,
		Deleted:        deleted,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create space output: %w", err)
	}
	return m, nil
}

// --- ListSpaces ---

// ListSpacesInput holds the parameters for listing spaces.
type ListSpacesInput struct {
	OperatorID       domain.AppUserID
	OrganizationName string `validate:"required"`
}

// NewListSpacesInput creates a validated ListSpacesInput.
func NewListSpacesInput(operatorID domain.AppUserID, organizationName string) (*ListSpacesInput, error) {
	if operatorID.IsZero() {
		return nil, errors.New("list spaces input operator id must not be zero")
	}
	m := &ListSpacesInput{
		OperatorID:       operatorID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list spaces input: %w", err)
	}
	return m, nil
}

// Item represents a single space in a list.
type Item struct {
	SpaceID        int
	OrganizationID domain.OrganizationID
	OwnerID        domain.AppUserID
	KeyName        string
	Name           string
	SpaceType      string
	Deleted        bool
}

// ListSpacesOutput holds the result of listing spaces.
type ListSpacesOutput struct {
	Spaces []Item
}
