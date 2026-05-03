package token

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// SessionToken represents a cookie-based session with sliding window expiry and absolute timeout.
type SessionToken struct {
	id               string
	userID           domain.AppUserID
	loginID          domain.LoginID
	organizationName string
	tokenHash        domain.TokenHash
	createdAt        time.Time
	expiresAt        time.Time
	revokedAt        *time.Time
}

// NewSessionToken creates a validated SessionToken.
func NewSessionToken(id string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, tokenHash domain.TokenHash, createdAt time.Time, expiresAt time.Time) (*SessionToken, error) {
	m := &SessionToken{
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
		return nil, fmt.Errorf("new session token: %w", err)
	}
	return m, nil
}

// ReconstructSessionToken reconstitutes a SessionToken from persistence (including RevokedAt).
func ReconstructSessionToken(id string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, tokenHash domain.TokenHash, createdAt time.Time, expiresAt time.Time, revokedAt *time.Time) *SessionToken {
	return &SessionToken{
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

func (t *SessionToken) validate() error {
	if t.id == "" {
		return fmt.Errorf("session token id is required: %w", domain.ErrInvalidArgument)
	}
	if t.userID.IsZero() {
		return fmt.Errorf("session token user id must not be zero: %w", domain.ErrInvalidArgument)
	}
	if t.loginID == "" {
		return fmt.Errorf("session token login id is required: %w", domain.ErrInvalidArgument)
	}
	if t.organizationName == "" {
		return fmt.Errorf("session token organization name is required: %w", domain.ErrInvalidArgument)
	}
	if len(t.tokenHash) != domain.TokenHashLength {
		return fmt.Errorf("session token hash must be 64 characters: %w", domain.ErrInvalidArgument)
	}
	if t.createdAt.IsZero() {
		return fmt.Errorf("session token created at is required: %w", domain.ErrInvalidArgument)
	}
	if t.expiresAt.IsZero() {
		return fmt.Errorf("session token expires at is required: %w", domain.ErrInvalidArgument)
	}
	return nil
}

// ID returns the token ID.
func (t *SessionToken) ID() string { return t.id }

// UserID returns the user ID.
func (t *SessionToken) UserID() domain.AppUserID { return t.userID }

// LoginID returns the login ID.
func (t *SessionToken) LoginID() domain.LoginID { return t.loginID }

// OrganizationName returns the organization name.
func (t *SessionToken) OrganizationName() string { return t.organizationName }

// TokenHash returns the SHA256 hash of the raw token.
func (t *SessionToken) TokenHash() domain.TokenHash { return t.tokenHash }

// CreatedAt returns the creation timestamp.
func (t *SessionToken) CreatedAt() time.Time { return t.createdAt }

// ExpiresAt returns the expiration timestamp.
func (t *SessionToken) ExpiresAt() time.Time { return t.expiresAt }

// RevokedAt returns a copy of the revocation timestamp. nil means the token is still active.
func (t *SessionToken) RevokedAt() *time.Time {
	if t.revokedAt == nil {
		return nil
	}
	copied := *t.revokedAt
	return &copied
}

// Revoke marks the token as revoked at the given time.
func (t *SessionToken) Revoke(now time.Time) {
	t.revokedAt = &now
}

// IsExpired returns true if the token's expiresAt is before now.
func (t *SessionToken) IsExpired(now time.Time) bool {
	return now.After(t.expiresAt)
}

// IsAbsoluteExpired returns true if the token has exceeded the absolute timeout from CreatedAt.
func (t *SessionToken) IsAbsoluteExpired(now time.Time, maxTTL time.Duration) bool {
	return now.After(t.createdAt.Add(maxTTL))
}

// IsRevoked returns true if the token has been revoked.
func (t *SessionToken) IsRevoked() bool {
	return t.revokedAt != nil
}

// IsValid returns true if the token is not revoked, not expired, and not absolute-expired.
func (t *SessionToken) IsValid(now time.Time, maxTTL time.Duration) bool {
	return !t.IsRevoked() && !t.IsExpired(now) && !t.IsAbsoluteExpired(now, maxTTL)
}

// ExtendExpiry extends the token's expiresAt using the sliding window TTL,
// capped by the absolute timeout (createdAt + maxTTL).
func (t *SessionToken) ExtendExpiry(now time.Time, slidingTTL time.Duration, maxTTL time.Duration) {
	newExpiry := now.Add(slidingTTL)
	absoluteMax := t.createdAt.Add(maxTTL)
	if newExpiry.After(absoluteMax) {
		t.expiresAt = absoluteMax
		return
	}
	t.expiresAt = newExpiry
}
