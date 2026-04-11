package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

func validRefreshTokenHash() domain.TokenHash {
	return token.HashToken("test-raw-token-for-refresh")
}

func validRefreshTokenArgs() (string, domain.AppUserID, domain.LoginID, string, domain.TokenHash, time.Time, time.Time) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return "rt-1", fixtureAppUserID, "user@example.com", "org1", validRefreshTokenHash(), now, now.Add(30 * 24 * time.Hour)
}

func Test_NewRefreshToken_shouldReturnToken_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()

	// when
	tk, err := token.NewRefreshToken(id, userID, loginID, org, hash, created, expires)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, tk.ID())
	assert.Equal(t, userID, tk.UserID())
	assert.Equal(t, loginID, tk.LoginID())
	assert.Equal(t, org, tk.OrganizationName())
	assert.Equal(t, hash, tk.TokenHash())
	assert.Equal(t, created, tk.CreatedAt())
	assert.Equal(t, expires, tk.ExpiresAt())
	assert.Nil(t, tk.RevokedAt())
}

func Test_NewRefreshToken_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()

	// when
	_, err := token.NewRefreshToken("", userID, loginID, org, hash, created, expires)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRefreshToken_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, org, hash, created, expires := validRefreshTokenArgs()

	// when
	_, err := token.NewRefreshToken(id, domain.AppUserID{}, loginID, org, hash, created, expires)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRefreshToken_shouldReturnError_whenTokenHashIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, _, created, expires := validRefreshTokenArgs()

	// when
	_, err := token.NewRefreshToken(id, userID, loginID, org, "short", created, expires)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_RefreshToken_Revoke_shouldSetRevokedAt_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	tk, _ := token.NewRefreshToken(id, userID, loginID, org, hash, created, expires)
	revokeTime := created.Add(10 * time.Minute)

	// when
	tk.Revoke(revokeTime)

	// then
	assert.True(t, tk.IsRevoked())
	assert.Equal(t, revokeTime, *tk.RevokedAt())
}

func Test_RefreshToken_RevokedAt_shouldReturnCopy_whenMutated(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := now.Add(10 * time.Minute)
	hash := validRefreshTokenHash()
	tk := token.ReconstructRefreshToken("rt-1", fixtureAppUserID, "user@example.com", "org1", hash, now, now.Add(30*24*time.Hour), &revokedAt)

	// when
	ptr := tk.RevokedAt()
	*ptr = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// then
	assert.Equal(t, revokedAt, *tk.RevokedAt())
}

func Test_RefreshToken_IsValid_shouldReturnTrue_whenNotRevokedAndNotExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	tk, _ := token.NewRefreshToken(id, userID, loginID, org, hash, created, expires)

	// when
	result := tk.IsValid(created.Add(time.Hour))

	// then
	assert.True(t, result)
}

func Test_RefreshToken_IsValid_shouldReturnFalse_whenRevoked(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	tk, _ := token.NewRefreshToken(id, userID, loginID, org, hash, created, expires)
	tk.Revoke(created.Add(time.Hour))

	// when
	result := tk.IsValid(created.Add(2 * time.Hour))

	// then
	assert.False(t, result)
}

func Test_RefreshToken_IsValid_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validRefreshTokenArgs()
	tk, _ := token.NewRefreshToken(id, userID, loginID, org, hash, created, expires)

	// when
	result := tk.IsValid(expires.Add(time.Second))

	// then
	assert.False(t, result)
}
