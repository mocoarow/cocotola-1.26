package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

func validSessionTokenHash() domain.TokenHash {
	return token.HashToken("test-raw-token-for-session")
}

func validSessionTokenArgs() (string, domain.AppUserID, domain.LoginID, string, domain.TokenHash, time.Time, time.Time) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return "st-1", fixtureAppUserID, "user@example.com", "org1", validSessionTokenHash(), now, now.Add(30 * time.Minute)
}

func Test_NewSessionToken_shouldReturnToken_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()

	// when
	tk, err := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)

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

func Test_NewSessionToken_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, userID, loginID, org, hash, created, expires := validSessionTokenArgs()

	// when
	_, err := token.NewSessionToken("", userID, loginID, org, hash, created, expires)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewSessionToken_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, org, hash, created, expires := validSessionTokenArgs()

	// when
	_, err := token.NewSessionToken(id, domain.AppUserID{}, loginID, org, hash, created, expires)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewSessionToken_shouldReturnError_whenTokenHashIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, _, created, expires := validSessionTokenArgs()

	// when
	_, err := token.NewSessionToken(id, userID, loginID, org, "short", created, expires)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_SessionToken_Revoke_shouldSetRevokedAt_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	revokeTime := created.Add(10 * time.Minute)

	// when
	tk.Revoke(revokeTime)

	// then
	assert.True(t, tk.IsRevoked())
	assert.Equal(t, revokeTime, *tk.RevokedAt())
}

func Test_SessionToken_RevokedAt_shouldReturnCopy_whenMutated(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := now.Add(10 * time.Minute)
	hash := validSessionTokenHash()
	tk := token.ReconstructSessionToken("st-1", fixtureAppUserID, "user@example.com", "org1", hash, now, now.Add(30*time.Minute), &revokedAt)

	// when
	ptr := tk.RevokedAt()
	*ptr = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// then
	assert.Equal(t, revokedAt, *tk.RevokedAt())
}

func Test_SessionToken_IsAbsoluteExpired_shouldReturnTrue_whenExceedsMaxTTL(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := tk.IsAbsoluteExpired(created.Add(maxTTL+time.Second), maxTTL)

	// then
	assert.True(t, result)
}

func Test_SessionToken_IsAbsoluteExpired_shouldReturnFalse_whenWithinMaxTTL(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := tk.IsAbsoluteExpired(created.Add(time.Hour), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_IsValid_shouldReturnTrue_whenAllChecksPass(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := tk.IsValid(created.Add(10*time.Minute), maxTTL)

	// then
	assert.True(t, result)
}

func Test_SessionToken_IsValid_shouldReturnFalse_whenRevoked(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	tk.Revoke(created.Add(5 * time.Minute))
	maxTTL := 24 * time.Hour

	// when
	result := tk.IsValid(created.Add(10*time.Minute), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_IsValid_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := tk.IsValid(expires.Add(time.Second), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_IsValid_shouldReturnFalse_whenAbsoluteExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, _ := validSessionTokenArgs()
	longExpiry := created.Add(48 * time.Hour)
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, longExpiry)
	maxTTL := 24 * time.Hour

	// when
	result := tk.IsValid(created.Add(maxTTL+time.Second), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_ExtendExpiry_shouldUpdateExpiresAt_whenWithinAbsoluteMax(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	slidingTTL := 30 * time.Minute
	maxTTL := 24 * time.Hour
	now := created.Add(10 * time.Minute)

	// when
	tk.ExtendExpiry(now, slidingTTL, maxTTL)

	// then
	assert.Equal(t, now.Add(slidingTTL), tk.ExpiresAt())
}

func Test_SessionToken_ExtendExpiry_shouldCapAtAbsoluteMax_whenSlidingExceedsMax(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	tk, _ := token.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	slidingTTL := 30 * time.Minute
	maxTTL := 24 * time.Hour
	now := created.Add(23*time.Hour + 50*time.Minute)

	// when
	tk.ExtendExpiry(now, slidingTTL, maxTTL)

	// then
	absoluteMax := created.Add(maxTTL)
	assert.Equal(t, absoluteMax, tk.ExpiresAt())
}
