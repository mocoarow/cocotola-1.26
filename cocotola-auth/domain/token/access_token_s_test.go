package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

var fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")

func validAccessTokenArgs() (string, string, domain.AppUserID, domain.LoginID, string, time.Time, time.Time) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return "jti-1", "rt-1", fixtureAppUserID, "user@example.com", "org1", now, now.Add(time.Hour)
}

func Test_NewAccessToken_shouldReturnToken_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()

	// when
	tk, err := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, tk.ID())
	assert.Equal(t, rtID, tk.RefreshTokenID())
	assert.Equal(t, userID, tk.UserID())
	assert.Equal(t, loginID, tk.LoginID())
	assert.Equal(t, org, tk.OrganizationName())
	assert.Equal(t, created, tk.CreatedAt())
	assert.Equal(t, expires, tk.ExpiresAt())
	assert.Nil(t, tk.RevokedAt())
}

func Test_NewAccessToken_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken("", rtID, userID, loginID, org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenRefreshTokenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, _, userID, loginID, org, created, expires := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken(id, "", userID, loginID, org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, _, loginID, org, created, expires := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken(id, rtID, domain.AppUserID{}, loginID, org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenLoginIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, _, org, created, expires := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken(id, rtID, userID, "", org, created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenOrganizationNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, _, created, expires := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken(id, rtID, userID, loginID, "", created, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenCreatedAtIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, _, expires := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken(id, rtID, userID, loginID, org, time.Time{}, expires)

	// then
	require.Error(t, err)
}

func Test_NewAccessToken_shouldReturnError_whenExpiresAtIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, _ := validAccessTokenArgs()

	// when
	_, err := token.NewAccessToken(id, rtID, userID, loginID, org, created, time.Time{})

	// then
	require.Error(t, err)
}

func Test_AccessToken_Revoke_shouldSetRevokedAt_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	tk, _ := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)
	revokeTime := created.Add(10 * time.Minute)

	// when
	tk.Revoke(revokeTime)

	// then
	assert.True(t, tk.IsRevoked())
	assert.NotNil(t, tk.RevokedAt())
	assert.Equal(t, revokeTime, *tk.RevokedAt())
}

func Test_AccessToken_RevokedAt_shouldReturnCopy_whenMutated(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := now.Add(10 * time.Minute)
	tk := token.ReconstructAccessToken("jti-1", "rt-1", fixtureAppUserID, "user@example.com", "org1", now, now.Add(time.Hour), &revokedAt)

	// when
	ptr := tk.RevokedAt()
	*ptr = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// then
	assert.Equal(t, revokedAt, *tk.RevokedAt())
}

func Test_AccessToken_IsExpired_shouldReturnTrue_whenNowIsAfterExpiresAt(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	tk, _ := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := tk.IsExpired(expires.Add(time.Second))

	// then
	assert.True(t, result)
}

func Test_AccessToken_IsExpired_shouldReturnFalse_whenNowIsBeforeExpiresAt(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	tk, _ := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := tk.IsExpired(created.Add(time.Minute))

	// then
	assert.False(t, result)
}

func Test_AccessToken_IsValid_shouldReturnTrue_whenNotRevokedAndNotExpired(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	tk, _ := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := tk.IsValid(created.Add(30 * time.Minute))

	// then
	assert.True(t, result)
}

func Test_AccessToken_IsValid_shouldReturnFalse_whenRevoked(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	tk, _ := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)
	tk.Revoke(created.Add(10 * time.Minute))

	// when
	result := tk.IsValid(created.Add(20 * time.Minute))

	// then
	assert.False(t, result)
}

func Test_AccessToken_IsValid_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	id, rtID, userID, loginID, org, created, expires := validAccessTokenArgs()
	tk, _ := token.NewAccessToken(id, rtID, userID, loginID, org, created, expires)

	// when
	result := tk.IsValid(expires.Add(time.Second))

	// then
	assert.False(t, result)
}
