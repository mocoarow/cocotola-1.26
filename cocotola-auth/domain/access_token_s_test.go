package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validAccessTokenArgs() (string, string, int, domain.LoginID, string, time.Time, time.Time) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return "jti-1", "rt-1", 1, "user@example.com", "org1", now, now.Add(time.Hour)
}

func Test_NewAccessToken_shouldReturnToken_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()

	// when
	token, err := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, token.ID())
	assert.Equal(t, rtID, token.RefreshTokenID())
	assert.Equal(t, userID, token.UserID())
	assert.Equal(t, loginID, token.LoginID())
	assert.Equal(t, org, token.OrganizationName())
	assert.Equal(t, created, token.CreatedAt())
	assert.Equal(t, expires, token.ExpiresAt())
	assert.Nil(t, token.RevokedAt())
}

func Test_NewAccessToken_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken("", rtID, userID, loginID, org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenRefreshTokenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, _, userID, loginID, org, created, expires := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken(id, "", userID, loginID, org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, _, loginID, org, created, expires := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken(id, rtID, 0, loginID, org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenLoginIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, _, org, created, expires := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken(id, rtID, userID, "", org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenOrganizationNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, _, created, expires := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken(id, rtID, userID, loginID, "", created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenCreatedAtIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, _, expires := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken(id, rtID, userID, loginID, org, time.Time{}, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenExpiresAtIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, _ := validAccessTokenArgs()

	// when
	_, err := domain.NewAccessToken(id, rtID, userID, loginID, org, created, time.Time{})

	// then
	require.Error(t, err)
}

func Test_AccessToken_Revoke_shouldSetRevokedAt_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	token, _ := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)
	revokeTime := created.Add(10 * time.Minute)

	// when
	token.Revoke(revokeTime)

	// then
	assert.True(t, token.IsRevoked())
	assert.NotNil(t, token.RevokedAt())
	assert.Equal(t, revokeTime, *token.RevokedAt())
}

func Test_AccessToken_RevokedAt_shouldReturnCopy_whenMutated(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := now.Add(10 * time.Minute)
	token := domain.ReconstructAccessToken("jti-1", "rt-1", 1, "user@example.com", "org1", now, now.Add(time.Hour), &revokedAt)

	// when
	ptr := token.RevokedAt()
	*ptr = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// then
	assert.Equal(t, revokedAt, *token.RevokedAt())
}

func Test_AccessToken_IsExpired_shouldReturnTrue_whenNowIsAfterExpiresAt(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	token, _ := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := token.IsExpired(expires.Add(time.Second))

	// then
	assert.True(t, result)
}

func Test_AccessToken_IsExpired_shouldReturnFalse_whenNowIsBeforeExpiresAt(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	token, _ := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := token.IsExpired(created.Add(time.Minute))

	// then
	assert.False(t, result)
}

func Test_AccessToken_IsValid_shouldReturnTrue_whenNotRevokedAndNotExpired(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	token, _ := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := token.IsValid(created.Add(30 * time.Minute))

	// then
	assert.True(t, result)
}

func Test_AccessToken_IsValid_shouldReturnFalse_whenRevoked(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	token, _ := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)
	token.Revoke(created.Add(10 * time.Minute))

	// when
	result := token.IsValid(created.Add(20 * time.Minute))

	// then
	assert.False(t, result)
}

func Test_AccessToken_IsValid_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	token, _ := domain.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := token.IsValid(expires.Add(time.Second))

	// then
	assert.False(t, result)
}
