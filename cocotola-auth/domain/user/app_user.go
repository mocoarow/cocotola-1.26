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
	provider       string
	providerID     string
	enabled        bool
}

// NewAppUser creates a validated AppUser for a brand-new aggregate (version 0).
func NewAppUser(id domain.AppUserID, organizationID domain.OrganizationID, loginID domain.LoginID, hashedPassword string, provider string, providerID string, enabled bool) (*AppUser, error) {
	m := &AppUser{
		id:             id,
		version:        0,
		organizationID: organizationID,
		loginID:        loginID,
		hashedPassword: hashedPassword,
		provider:       provider,
		providerID:     providerID,
		enabled:        enabled,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructAppUser reconstitutes an AppUser from persistence.
// The returned aggregate has version 0; callers that load from storage should
// call WithVersion to set the persisted version so Save can perform an
// optimistic-lock compare-and-swap.
func ReconstructAppUser(id domain.AppUserID, organizationID domain.OrganizationID, loginID domain.LoginID, hashedPassword string, provider string, providerID string, enabled bool) *AppUser {
	return &AppUser{
		id:             id,
		version:        0,
		organizationID: organizationID,
		loginID:        loginID,
		hashedPassword: hashedPassword,
		provider:       provider,
		providerID:     providerID,
		enabled:        enabled,
	}
}

// WithVersion sets the persisted row version on a reconstituted aggregate.
// Only the gateway/repository layer should call this when loading from storage.
func (u *AppUser) WithVersion(version int) *AppUser {
	u.version = version
	return u
}

// Version returns the persisted row version (0 = new, not yet saved).
func (u *AppUser) Version() int { return u.version }

// IncrementVersion bumps the version after a successful persist. Callers in the
// repository layer invoke this after a successful INSERT or UPDATE so that a
// subsequent Save on the same aggregate uses the correct compare-and-swap target.
func (u *AppUser) IncrementVersion() { u.version++ }

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
	if (u.provider == "") != (u.providerID == "") {
		return errors.New("app user provider and provider id must both be set or both be empty")
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

// Provider returns the external identity provider name (empty if not linked).
func (u *AppUser) Provider() string { return u.provider }

// ProviderID returns the external identity provider user ID (empty if not linked).
func (u *AppUser) ProviderID() string { return u.providerID }

// IsLinkedToProvider reports whether the user has been linked to an external identity provider.
func (u *AppUser) IsLinkedToProvider() bool {
	return u.provider != "" && u.providerID != ""
}

// Enabled returns whether the user is enabled.
func (u *AppUser) Enabled() bool { return u.enabled }

// Enable enables the user.
func (u *AppUser) Enable() { u.enabled = true }

// Disable disables the user.
func (u *AppUser) Disable() { u.enabled = false }

// LinkProvider associates an external identity provider with this user.
// Returns ErrAppUserAlreadyLinked if the user is already linked to a provider,
// preventing silent overwrites that could enable account takeover.
func (u *AppUser) LinkProvider(provider string, providerID string) error {
	if provider == "" || providerID == "" {
		return errors.New("provider and provider id are required")
	}
	if u.IsLinkedToProvider() {
		return domain.ErrAppUserAlreadyLinked
	}
	u.provider = provider
	u.providerID = providerID
	return nil
}

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
