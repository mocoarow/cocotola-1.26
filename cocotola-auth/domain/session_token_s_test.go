package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validSessionTokenHash() domain.TokenHash {
	return domain.HashToken("test-raw-token-for-session")
}

func validSessionTokenArgs() (string, int, domain.LoginID, string, domain.TokenHash, time.Time, time.Time) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return "st-1", 1, "user@example.com", "org1", validSessionTokenHash(), now, now.Add(30 * time.Minute)
}

func Test_NewSessionToken_shouldReturnToken_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()

	// when
	token, err := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, token.ID())
	assert.Equal(t, userID, token.UserID())
	assert.Equal(t, loginID, token.LoginID())
	assert.Equal(t, org, token.OrganizationName())
	assert.Equal(t, hash, token.TokenHash())
	assert.Equal(t, created, token.CreatedAt())
	assert.Equal(t, expires, token.ExpiresAt())
	assert.Nil(t, token.RevokedAt())
}

func Test_NewSessionToken_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, userID, loginID, org, hash, created, expires := validSessionTokenArgs()

	// when
	_, err := domain.NewSessionToken("", userID, loginID, org, hash, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewSessionToken_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, org, hash, created, expires := validSessionTokenArgs()

	// when
	_, err := domain.NewSessionToken(id, 0, loginID, org, hash, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewSessionToken_shouldReturnError_whenTokenHashIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, _, created, expires := validSessionTokenArgs()

	// when
	_, err := domain.NewSessionToken(id, userID, loginID, org, "short", created, expires)

	// then
	require.Error(t, err)
}

func Test_SessionToken_Revoke_shouldSetRevokedAt_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	revokeTime := created.Add(10 * time.Minute)

	// when
	token.Revoke(revokeTime)

	// then
	assert.True(t, token.IsRevoked())
	assert.Equal(t, revokeTime, *token.RevokedAt())
}

func Test_SessionToken_RevokedAt_shouldReturnCopy_whenMutated(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := now.Add(10 * time.Minute)
	hash := validSessionTokenHash()
	token := domain.ReconstructSessionToken("st-1", 1, "user@example.com", "org1", hash, now, now.Add(30*time.Minute), &revokedAt)

	// when
	ptr := token.RevokedAt()
	*ptr = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// then
	assert.Equal(t, revokedAt, *token.RevokedAt())
}

func Test_SessionToken_IsAbsoluteExpired_shouldReturnTrue_whenExceedsMaxTTL(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := token.IsAbsoluteExpired(created.Add(maxTTL+time.Second), maxTTL)

	// then
	assert.True(t, result)
}

func Test_SessionToken_IsAbsoluteExpired_shouldReturnFalse_whenWithinMaxTTL(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := token.IsAbsoluteExpired(created.Add(time.Hour), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_IsValid_shouldReturnTrue_whenAllChecksPass(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := token.IsValid(created.Add(10*time.Minute), maxTTL)

	// then
	assert.True(t, result)
}

func Test_SessionToken_IsValid_shouldReturnFalse_whenRevoked(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	token.Revoke(created.Add(5 * time.Minute))
	maxTTL := 24 * time.Hour

	// when
	result := token.IsValid(created.Add(10*time.Minute), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_IsValid_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	maxTTL := 24 * time.Hour

	// when
	result := token.IsValid(expires.Add(time.Second), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_IsValid_shouldReturnFalse_whenAbsoluteExpired(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, _ := validSessionTokenArgs()
	longExpiry := created.Add(48 * time.Hour)
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, longExpiry)
	maxTTL := 24 * time.Hour

	// when
	result := token.IsValid(created.Add(maxTTL+time.Second), maxTTL)

	// then
	assert.False(t, result)
}

func Test_SessionToken_ExtendExpiry_shouldUpdateExpiresAt_whenWithinAbsoluteMax(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	slidingTTL := 30 * time.Minute
	maxTTL := 24 * time.Hour
	now := created.Add(10 * time.Minute)

	// when
	token.ExtendExpiry(now, slidingTTL, maxTTL)

	// then
	assert.Equal(t, now.Add(slidingTTL), token.ExpiresAt())
}

func Test_SessionToken_ExtendExpiry_shouldCapAtAbsoluteMax_whenSlidingExceedsMax(t *testing.T) {
	t.Parallel()

	// given
	id, userID, loginID, org, hash, created, expires := validSessionTokenArgs()
	token, _ := domain.NewSessionToken(id, userID, loginID, org, hash, created, expires)
	slidingTTL := 30 * time.Minute
	maxTTL := 24 * time.Hour
	now := created.Add(23*time.Hour + 50*time.Minute)

	// when
	token.ExtendExpiry(now, slidingTTL, maxTTL)

	// then
	absoluteMax := created.Add(maxTTL)
	assert.Equal(t, absoluteMax, token.ExpiresAt())
}
