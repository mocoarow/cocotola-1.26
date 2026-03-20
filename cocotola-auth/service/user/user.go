// Package user provides service-layer types for user management input/output validation.
package user

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// --- CreateAppUser ---

// CreateAppUserInput holds the parameters for creating an app user.
type CreateAppUserInput struct {
	OrganizationID int    `validate:"required,gt=0"`
	LoginID        string `validate:"required"`
	Password       string `validate:"required,min=8"`
}

// NewCreateAppUserInput creates a validated CreateAppUserInput.
func NewCreateAppUserInput(organizationID int, loginID string, password string) (*CreateAppUserInput, error) {
	m := &CreateAppUserInput{
		OrganizationID: organizationID,
		LoginID:        loginID,
		Password:       password,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create app user input: %w", err)
	}
	return m, nil
}

// CreateAppUserOutput holds the result of creating an app user.
type CreateAppUserOutput struct {
	AppUserID      int    `validate:"required,gt=0"`
	OrganizationID int    `validate:"required,gt=0"`
	LoginID        string `validate:"required"`
	Enabled        bool
}

// NewCreateAppUserOutput creates a validated CreateAppUserOutput.
func NewCreateAppUserOutput(appUserID int, organizationID int, loginID string, enabled bool) (*CreateAppUserOutput, error) {
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
	AppUserID   int    `validate:"required,gt=0"`
	NewPassword string `validate:"required,min=8"`
}

// NewChangePasswordInput creates a validated ChangePasswordInput.
func NewChangePasswordInput(appUserID int, newPassword string) (*ChangePasswordInput, error) {
	m := &ChangePasswordInput{
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
	AppUserID int `validate:"required,gt=0"`
}

// NewChangePasswordOutput creates a validated ChangePasswordOutput.
func NewChangePasswordOutput(appUserID int) (*ChangePasswordOutput, error) {
	m := &ChangePasswordOutput{
		AppUserID: appUserID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate change password output: %w", err)
	}
	return m, nil
}
