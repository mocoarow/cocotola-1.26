// Package token contains the token aggregates (access, refresh, session) of the cocotola-auth domain.
package token

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// AccessToken represents a JWT access token registered in the whitelist.
// The ID is the JWT's JTI claim. No TokenHash is needed since JTI is used for lookup.
type AccessToken struct {
	id               string
	refreshTokenID   string
	userID           domain.AppUserID
	loginID          domain.LoginID
	organizationName string
	createdAt        time.Time
	expiresAt        time.Time
	revokedAt        *time.Time
}

// NewAccessToken creates a validated AccessToken.
func NewAccessToken(id string, refreshTokenID string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, createdAt time.Time, expiresAt time.Time) (*AccessToken, error) {
	m := &AccessToken{
		id:               id,
		refreshTokenID:   refreshTokenID,
		userID:           userID,
		loginID:          loginID,
		organizationName: organizationName,
		createdAt:        createdAt,
		expiresAt:        expiresAt,
		revokedAt:        nil,
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("new access token: %w", err)
	}
	return m, nil
}

// ReconstructAccessToken reconstitutes an AccessToken from persistence (including RevokedAt).
func ReconstructAccessToken(id string, refreshTokenID string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, createdAt time.Time, expiresAt time.Time, revokedAt *time.Time) *AccessToken {
	return &AccessToken{
		id:               id,
		refreshTokenID:   refreshTokenID,
		userID:           userID,
		loginID:          loginID,
		organizationName: organizationName,
		createdAt:        createdAt,
		expiresAt:        expiresAt,
		revokedAt:        revokedAt,
	}
}

func (t *AccessToken) validate() error {
	if t.id == "" {
		return fmt.Errorf("access token id is required: %w", domain.ErrInvalidArgument)
	}
	if t.refreshTokenID == "" {
		return fmt.Errorf("access token refresh token id is required: %w", domain.ErrInvalidArgument)
	}
	if t.userID.IsZero() {
		return fmt.Errorf("access token user id is required: %w", domain.ErrInvalidArgument)
	}
	if t.loginID == "" {
		return fmt.Errorf("access token login id is required: %w", domain.ErrInvalidArgument)
	}
	if t.organizationName == "" {
		return fmt.Errorf("access token organization name is required: %w", domain.ErrInvalidArgument)
	}
	if t.createdAt.IsZero() {
		return fmt.Errorf("access token created at is required: %w", domain.ErrInvalidArgument)
	}
	if t.expiresAt.IsZero() {
		return fmt.Errorf("access token expires at is required: %w", domain.ErrInvalidArgument)
	}
	return nil
}

// ID returns the token ID (= JWT JTI).
func (t *AccessToken) ID() string { return t.id }

// RefreshTokenID returns the associated refresh token ID.
func (t *AccessToken) RefreshTokenID() string { return t.refreshTokenID }

// UserID returns the user ID.
func (t *AccessToken) UserID() domain.AppUserID { return t.userID }

// LoginID returns the login ID.
func (t *AccessToken) LoginID() domain.LoginID { return t.loginID }

// OrganizationName returns the organization name.
func (t *AccessToken) OrganizationName() string { return t.organizationName }

// CreatedAt returns the creation timestamp.
func (t *AccessToken) CreatedAt() time.Time { return t.createdAt }

// ExpiresAt returns the expiration timestamp.
func (t *AccessToken) ExpiresAt() time.Time { return t.expiresAt }

// RevokedAt returns a copy of the revocation timestamp. nil means the token is still active.
func (t *AccessToken) RevokedAt() *time.Time {
	if t.revokedAt == nil {
		return nil
	}
	copied := *t.revokedAt
	return &copied
}

// Revoke marks the token as revoked at the given time.
func (t *AccessToken) Revoke(now time.Time) {
	t.revokedAt = &now
}

// IsExpired returns true if the token's expiresAt is before now.
func (t *AccessToken) IsExpired(now time.Time) bool {
	return now.After(t.expiresAt)
}

// IsRevoked returns true if the token has been revoked.
func (t *AccessToken) IsRevoked() bool {
	return t.revokedAt != nil
}

// IsValid returns true if the token is not revoked and not expired.
func (t *AccessToken) IsValid(now time.Time) bool {
	return !t.IsRevoked() && !t.IsExpired(now)
}
