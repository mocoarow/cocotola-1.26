package token

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// RefreshToken represents a long-lived opaque token used for token-based authentication.
type RefreshToken struct {
	id               string
	userID           domain.AppUserID
	loginID          domain.LoginID
	organizationName string
	tokenHash        domain.TokenHash
	createdAt        time.Time
	expiresAt        time.Time
	revokedAt        *time.Time
}

// NewRefreshToken creates a validated RefreshToken.
func NewRefreshToken(id string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, tokenHash domain.TokenHash, createdAt time.Time, expiresAt time.Time) (*RefreshToken, error) {
	m := &RefreshToken{
		id:               id,
		userID:           userID,
		loginID:          loginID,
		organizationName: organizationName,
		tokenHash:        tokenHash,
		createdAt:        createdAt,
		expiresAt:        expiresAt,
		revokedAt:        nil,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructRefreshToken reconstitutes a RefreshToken from persistence (including RevokedAt).
func ReconstructRefreshToken(id string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, tokenHash domain.TokenHash, createdAt time.Time, expiresAt time.Time, revokedAt *time.Time) *RefreshToken {
	return &RefreshToken{
		id:               id,
		userID:           userID,
		loginID:          loginID,
		organizationName: organizationName,
		tokenHash:        tokenHash,
		createdAt:        createdAt,
		expiresAt:        expiresAt,
		revokedAt:        revokedAt,
	}
}

func (t *RefreshToken) validate() error {
	if t.id == "" {
		return fmt.Errorf("refresh token id is required: %w", domain.ErrInvalidArgument)
	}
	if t.userID.IsZero() {
		return fmt.Errorf("refresh token user id is required: %w", domain.ErrInvalidArgument)
	}
	if t.loginID == "" {
		return fmt.Errorf("refresh token login id is required: %w", domain.ErrInvalidArgument)
	}
	if t.organizationName == "" {
		return fmt.Errorf("refresh token organization name is required: %w", domain.ErrInvalidArgument)
	}
	if len(t.tokenHash) != domain.TokenHashLength {
		return fmt.Errorf("refresh token hash must be 64 characters: %w", domain.ErrInvalidArgument)
	}
	if t.createdAt.IsZero() {
		return fmt.Errorf("refresh token created at is required: %w", domain.ErrInvalidArgument)
	}
	if t.expiresAt.IsZero() {
		return fmt.Errorf("refresh token expires at is required: %w", domain.ErrInvalidArgument)
	}
	return nil
}

// ID returns the token ID.
func (t *RefreshToken) ID() string { return t.id }

// UserID returns the user ID.
func (t *RefreshToken) UserID() domain.AppUserID { return t.userID }

// LoginID returns the login ID.
func (t *RefreshToken) LoginID() domain.LoginID { return t.loginID }

// OrganizationName returns the organization name.
func (t *RefreshToken) OrganizationName() string { return t.organizationName }

// TokenHash returns the SHA256 hash of the raw token.
func (t *RefreshToken) TokenHash() domain.TokenHash { return t.tokenHash }

// CreatedAt returns the creation timestamp.
func (t *RefreshToken) CreatedAt() time.Time { return t.createdAt }

// ExpiresAt returns the expiration timestamp.
func (t *RefreshToken) ExpiresAt() time.Time { return t.expiresAt }

// RevokedAt returns a copy of the revocation timestamp. nil means the token is still active.
func (t *RefreshToken) RevokedAt() *time.Time {
	if t.revokedAt == nil {
		return nil
	}
	copied := *t.revokedAt
	return &copied
}

// Revoke marks the token as revoked at the given time.
func (t *RefreshToken) Revoke(now time.Time) {
	t.revokedAt = &now
}

// IsExpired returns true if the token's expiresAt is before now.
func (t *RefreshToken) IsExpired(now time.Time) bool {
	return now.After(t.expiresAt)
}

// IsRevoked returns true if the token has been revoked.
func (t *RefreshToken) IsRevoked() bool {
	return t.revokedAt != nil
}

// IsValid returns true if the token is not revoked and not expired.
func (t *RefreshToken) IsValid(now time.Time) bool {
	return !t.IsRevoked() && !t.IsExpired(now)
}
