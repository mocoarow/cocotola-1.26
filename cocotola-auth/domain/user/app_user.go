// Package user contains the user aggregate of the cocotola-auth domain.
package user

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// AppUser represents a user belonging to an organization.
type AppUser struct {
	id             domain.AppUserID
	version        int
	organizationID domain.OrganizationID
	loginID        domain.LoginID
	hashedPassword string
	enabled        bool
}

// NewAppUser creates a validated AppUser for a brand-new aggregate (version 0).
func NewAppUser(id domain.AppUserID, organizationID domain.OrganizationID, loginID domain.LoginID, hashedPassword string, enabled bool) (*AppUser, error) {
	m := &AppUser{
		id:             id,
		version:        0,
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
// Callers that load from storage must call WithVersion to set the persisted
// version so Save can perform an optimistic-lock compare-and-swap.
func ReconstructAppUser(id domain.AppUserID, organizationID domain.OrganizationID, loginID domain.LoginID, hashedPassword string, enabled bool) *AppUser {
	return &AppUser{
		id:             id,
		version:        0,
		organizationID: organizationID,
		loginID:        loginID,
		hashedPassword: hashedPassword,
		enabled:        enabled,
	}
}

// WithVersion sets the persisted row version on a reconstituted aggregate.
// Only the gateway/repository layer should call this when loading from storage.
func (u *AppUser) WithVersion(version int) *AppUser {
	u.version = version
	return u
}

// Version returns the aggregate version (0 = new, not yet saved).
func (u *AppUser) Version() int { return u.version }

func (u *AppUser) validate() error {
	if u.id.IsZero() {
		return errors.New("app user id must not be zero")
	}
	if u.organizationID.IsZero() {
		return errors.New("app user organization id must not be zero")
	}
	if u.loginID == "" {
		return errors.New("app user login id is required")
	}
	return nil
}

// ID returns the user ID.
func (u *AppUser) ID() domain.AppUserID { return u.id }

// OrganizationID returns the organization ID.
func (u *AppUser) OrganizationID() domain.OrganizationID { return u.organizationID }

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
