// Package user contains the user aggregate of the cocotola-auth domain.
package user

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// AppUser represents a user belonging to an organization.
type AppUser struct {
	id             int
	organizationID int
	loginID        domain.LoginID
	hashedPassword string
	enabled        bool
}

// NewAppUser creates a validated AppUser.
func NewAppUser(id int, organizationID int, loginID domain.LoginID, hashedPassword string, enabled bool) (*AppUser, error) {
	m := &AppUser{
		id:             id,
		organizationID: organizationID,
		loginID:        loginID,
		hashedPassword: hashedPassword,
		enabled:        enabled,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructAppUser reconstitutes an AppUser from persistence.
func ReconstructAppUser(id int, organizationID int, loginID domain.LoginID, hashedPassword string, enabled bool) *AppUser {
	return &AppUser{
		id:             id,
		organizationID: organizationID,
		loginID:        loginID,
		hashedPassword: hashedPassword,
		enabled:        enabled,
	}
}

func (u *AppUser) validate() error {
	if u.id <= 0 {
		return errors.New("app user id must be positive")
	}
	if u.organizationID <= 0 {
		return errors.New("app user organization id must be positive")
	}
	if u.loginID == "" {
		return errors.New("app user login id is required")
	}
	return nil
}

// ID returns the user ID.
func (u *AppUser) ID() int { return u.id }

// OrganizationID returns the organization ID.
func (u *AppUser) OrganizationID() int { return u.organizationID }

// LoginID returns the login ID.
func (u *AppUser) LoginID() domain.LoginID { return u.loginID }

// HashedPassword returns the bcrypt-hashed password.
func (u *AppUser) HashedPassword() string { return u.hashedPassword }

// Enabled returns whether the user is enabled.
func (u *AppUser) Enabled() bool { return u.enabled }

// Enable enables the user.
func (u *AppUser) Enable() { u.enabled = true }

// Disable disables the user.
func (u *AppUser) Disable() { u.enabled = false }

// ChangePassword validates the raw password against the policy, hashes it, and updates the user.
func (u *AppUser) ChangePassword(rawPassword string, hasher PasswordHasher) error {
	hashed, err := HashPassword(rawPassword, hasher)
	if err != nil {
		return err
	}
	u.hashedPassword = hashed
	return nil
}

// VerifyPassword checks the raw password against the stored hash.
func (u *AppUser) VerifyPassword(rawPassword string, hasher PasswordHasher) error {
	if err := hasher.Compare(u.hashedPassword, rawPassword); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}
