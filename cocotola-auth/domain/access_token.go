package domain

import (
	"errors"
	"time"
)

// AccessToken represents a JWT access token registered in the whitelist.
// The ID is the JWT's JTI claim. No TokenHash is needed since JTI is used for lookup.
type AccessToken struct {
	id               string
	refreshTokenID   string
	userID           int
	loginID          LoginID
	organizationName string
	createdAt        time.Time
	expiresAt        time.Time
	revokedAt        *time.Time
}

// NewAccessToken creates a validated AccessToken.
func NewAccessToken(id string, refreshTokenID string, userID int, loginID LoginID, organizationName string, createdAt time.Time, expiresAt time.Time) (*AccessToken, error) {
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
		return nil, err
	}
	return m, nil
}

// ReconstructAccessToken reconstitutes an AccessToken from persistence (including RevokedAt).
func ReconstructAccessToken(id string, refreshTokenID string, userID int, loginID LoginID, organizationName string, createdAt time.Time, expiresAt time.Time, revokedAt *time.Time) *AccessToken {
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
		return errors.New("access token id is required")
	}
	if t.refreshTokenID == "" {
		return errors.New("access token refresh token id is required")
	}
	if t.userID <= 0 {
		return errors.New("access token user id must be positive")
	}
	if t.loginID == "" {
		return errors.New("access token login id is required")
	}
	if t.organizationName == "" {
		return errors.New("access token organization name is required")
	}
	if t.createdAt.IsZero() {
		return errors.New("access token created at is required")
	}
	if t.expiresAt.IsZero() {
		return errors.New("access token expires at is required")
	}
	return nil
}

// ID returns the token ID (= JWT JTI).
func (t *AccessToken) ID() string { return t.id }

// RefreshTokenID returns the associated refresh token ID.
func (t *AccessToken) RefreshTokenID() string { return t.refreshTokenID }

// UserID returns the user ID.
func (t *AccessToken) UserID() int { return t.userID }

// LoginID returns the login ID.
func (t *AccessToken) LoginID() LoginID { return t.loginID }

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
