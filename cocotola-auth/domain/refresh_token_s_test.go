package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validRefreshTokenHash() domain.TokenHash {
	return domain.HashToken("test-raw-token-for-refresh")
}

func validRefreshTokenArgs() (string, int, domain.LoginID, string, domain.TokenHash, time.Time, time.Time) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return "rt-1", 1, "user@example.com", "org1", validRefreshTokenHash(), now, now.Add(30 * 24 * time.Hour)
}

func Test_NewRefreshToken_shouldReturnToken_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()

	// when
	token, err := domain.NewRefreshToken(id, userID, loginID, org, hash, created, expires)

	// then
	assert.NoError(t, err)
	assert.Equal(t, id, token.ID())
	assert.Equal(t, userID, token.UserID())
	assert.Equal(t, loginID, token.LoginID())
	assert.Equal(t, org, token.OrganizationName())
	assert.Equal(t, hash, token.TokenHash())
	assert.Equal(t, created, token.CreatedAt())
	assert.Equal(t, expires, token.ExpiresAt())
	assert.Nil(t, token.RevokedAt())
}

func Test_NewRefreshToken_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()

	// when
	_, err := domain.NewRefreshToken("", userID, loginID, org, hash, created, expires)

	// then
	assert.Error(t, err)
}

func Test_NewRefreshToken_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, org, hash, created, expires := validRefreshTokenArgs()

	// when
	_, err := domain.NewRefreshToken(id, 0, loginID, org, hash, created, expires)

	// then
	assert.Error(t, err)
}

func Test_NewRefreshToken_shouldReturnError_whenTokenHashIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, _, created, expires := validRefreshTokenArgs()

	// when
	_, err := domain.NewRefreshToken(id, userID, loginID, org, "short", created, expires)

	// then
	assert.Error(t, err)
}

func Test_RefreshToken_Revoke_shouldSetRevokedAt_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	token, _ := domain.NewRefreshToken(id, userID, loginID, org, hash, created, expires)
	revokeTime := created.Add(10 * time.Minute)

	// when
	token.Revoke(revokeTime)

	// then
	assert.True(t, token.IsRevoked())
	assert.Equal(t, revokeTime, *token.RevokedAt())
}

func Test_RefreshToken_RevokedAt_shouldReturnCopy_whenMutated(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := now.Add(10 * time.Minute)
	hash := validRefreshTokenHash()
	token := domain.ReconstructRefreshToken("rt-1", 1, "user@example.com", "org1", hash, now, now.Add(30*24*time.Hour), &revokedAt)

	// when
	ptr := token.RevokedAt()
	*ptr = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// then
	assert.Equal(t, revokedAt, *token.RevokedAt())
}

func Test_RefreshToken_IsValid_shouldReturnTrue_whenNotRevokedAndNotExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	token, _ := domain.NewRefreshToken(id, userID, loginID, org, hash, created, expires)

	// when
	result := token.IsValid(created.Add(time.Hour))

	// then
	assert.True(t, result)
}

func Test_RefreshToken_IsValid_shouldReturnFalse_whenRevoked(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	token, _ := domain.NewRefreshToken(id, userID, loginID, org, hash, created, expires)
	token.Revoke(created.Add(time.Hour))

	// when
	result := token.IsValid(created.Add(2 * time.Hour))

	// then
	assert.False(t, result)
}

func Test_RefreshToken_IsValid_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	token, _ := domain.NewRefreshToken(id, userID, loginID, org, hash, created, expires)

	// when
	result := token.IsValid(expires.Add(time.Second))

	// then
	assert.False(t, result)
}
