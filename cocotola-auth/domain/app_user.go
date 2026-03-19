package domain

import "errors"

// AppUser represents a user belonging to an organization.
type AppUser struct {
	id             int
	organizationID int
	loginID        LoginID
	enabled        bool
}

// NewAppUser creates a validated AppUser.
func NewAppUser(id int, organizationID int, loginID LoginID, enabled bool) (*AppUser, error) {
	m := &AppUser{
		id:             id,
		organizationID: organizationID,
		loginID:        loginID,
		enabled:        enabled,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructAppUser reconstitutes an AppUser from persistence.
func ReconstructAppUser(id int, organizationID int, loginID LoginID, enabled bool) *AppUser {
	return &AppUser{
		id:             id,
		organizationID: organizationID,
		loginID:        loginID,
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
func (u *AppUser) LoginID() LoginID { return u.loginID }

// Enabled returns whether the user is enabled.
func (u *AppUser) Enabled() bool { return u.enabled }

// Enable enables the user.
func (u *AppUser) Enable() { u.enabled = true }

// Disable disables the user.
func (u *AppUser) Disable() { u.enabled = false }
