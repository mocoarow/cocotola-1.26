// Package user provides service-layer types for user management input/output validation.
package user

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// --- CreateAppUser ---

// CreateAppUserInput holds the parameters for creating an app user.
type CreateAppUserInput struct {
	OperatorID       domain.AppUserID
	OrganizationName string `validate:"required"`
	LoginID          string `validate:"required"`
	Password         string `validate:"required,min=8"`
}

// NewCreateAppUserInput creates a validated CreateAppUserInput.
func NewCreateAppUserInput(operatorID domain.AppUserID, organizationName string, loginID string, password string) (*CreateAppUserInput, error) {
	if operatorID.IsZero() {
		return nil, errors.New("create app user input operator id must not be zero")
	}
	m := &CreateAppUserInput{
		OperatorID:       operatorID,
		OrganizationName: organizationName,
		LoginID:          loginID,
		Password:         password,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create app user input: %w", err)
	}
	return m, nil
}

// CreateAppUserOutput holds the result of creating an app user.
type CreateAppUserOutput struct {
	AppUserID      domain.AppUserID
	OrganizationID domain.OrganizationID
	LoginID        string `validate:"required"`
	Enabled        bool
}

// NewCreateAppUserOutput creates a validated CreateAppUserOutput.
func NewCreateAppUserOutput(appUserID domain.AppUserID, organizationID domain.OrganizationID, loginID string, enabled bool) (*CreateAppUserOutput, error) {
	if appUserID.IsZero() {
		return nil, errors.New("create app user output app user id must not be zero")
	}
	if organizationID.IsZero() {
		return nil, errors.New("create app user output organization id must not be zero")
	}
	m := &CreateAppUserOutput{
		AppUserID:      appUserID,
		OrganizationID: organizationID,
		LoginID:        loginID,
		Enabled:        enabled,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create app user output: %w", err)
	}
	return m, nil
}

// --- ChangePassword ---

// ChangePasswordInput holds the parameters for changing a user's password.
type ChangePasswordInput struct {
	OperatorID  domain.AppUserID
	AppUserID   domain.AppUserID
	NewPassword string `validate:"required,min=8"`
}

// NewChangePasswordInput creates a validated ChangePasswordInput.
func NewChangePasswordInput(operatorID domain.AppUserID, appUserID domain.AppUserID, newPassword string) (*ChangePasswordInput, error) {
	if operatorID.IsZero() {
		return nil, errors.New("change password input operator id must not be zero")
	}
	if appUserID.IsZero() {
		return nil, errors.New("change password input app user id must not be zero")
	}
	m := &ChangePasswordInput{
		OperatorID:  operatorID,
		AppUserID:   appUserID,
		NewPassword: newPassword,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate change password input: %w", err)
	}
	return m, nil
}

// ChangePasswordOutput holds the result of changing a user's password.
type ChangePasswordOutput struct {
	AppUserID domain.AppUserID
}

// NewChangePasswordOutput creates a validated ChangePasswordOutput.
func NewChangePasswordOutput(appUserID domain.AppUserID) (*ChangePasswordOutput, error) {
	if appUserID.IsZero() {
		return nil, errors.New("change password output app user id must not be zero")
	}
	m := &ChangePasswordOutput{
		AppUserID: appUserID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate change password output: %w", err)
	}
	return m, nil
}
