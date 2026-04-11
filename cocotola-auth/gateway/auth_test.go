package gateway_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func newTestJWTManager(t *testing.T) *gateway.JWTManager {
	t.Helper()
	signingKey := []byte("test-signing-key-that-is-long-enough-for-hmac")
	return gateway.NewJWTManager(signingKey, jwt.SigningMethodHS256, 60*time.Minute)
}

func Test_JWTManager_CreateAccessToken_shouldReturnToken_whenValidInput(t *testing.T) {
	t.Parallel()

	// given
	m := newTestJWTManager(t)

	// when
	token, err := m.CreateAccessToken("user1", fixtureAppUserID, "org1", "test-jti-123")

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func Test_JWTManager_ParseAccessToken_shouldReturnUserInfoAndJTI_whenValidToken(t *testing.T) {
	t.Parallel()

	// given
	m := newTestJWTManager(t)
	token, err := m.CreateAccessToken("user1", fixtureAppUserID, "org1", "test-jti-456")
	require.NoError(t, err)

	// when
	userInfo, jti, err := m.ParseAccessToken(token)

	// then
	require.NoError(t, err)
	require.NotNil(t, userInfo)
	assert.True(t, fixtureAppUserID.Equal(userInfo.UserID))
	assert.Equal(t, "user1", userInfo.LoginID)
	assert.Equal(t, "org1", userInfo.OrganizationName)
	assert.Equal(t, "test-jti-456", jti)
}

func Test_JWTManager_ParseAccessToken_shouldReturnError_whenTokenIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	m := newTestJWTManager(t)

	// when
	userInfo, jti, err := m.ParseAccessToken("invalid-token-string")

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
	assert.Nil(t, userInfo)
	assert.Empty(t, jti)
}

func Test_JWTManager_ParseAccessToken_shouldReturnError_whenTokenIsSignedWithDifferentKey(t *testing.T) {
	t.Parallel()

	// given
	creator := gateway.NewJWTManager([]byte("original-key-that-is-long-enough"), jwt.SigningMethodHS256, 60*time.Minute)
	parser := gateway.NewJWTManager([]byte("different-key-that-is-long-enough"), jwt.SigningMethodHS256, 60*time.Minute)
	token, err := creator.CreateAccessToken("user1", fixtureAppUserID, "org1", "jti-1")
	require.NoError(t, err)

	// when
	userInfo, jti, err := parser.ParseAccessToken(token)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
	assert.Nil(t, userInfo)
	assert.Empty(t, jti)
}

func Test_JWTManager_ParseAccessToken_shouldReturnError_whenTokenIsExpired(t *testing.T) {
	t.Parallel()

	// given
	m := gateway.NewJWTManager([]byte("test-signing-key-that-is-long-enough-for-hmac"), jwt.SigningMethodHS256, -1*time.Minute)
	token, err := m.CreateAccessToken("user1", fixtureAppUserID, "org1", "jti-expired")
	require.NoError(t, err)

	// when
	userInfo, jti, err := m.ParseAccessToken(token)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
	assert.Nil(t, userInfo)
	assert.Empty(t, jti)
}

func Test_JWTManager_ParseAccessToken_shouldReturnExpiresAt(t *testing.T) {
	t.Parallel()

	// given
	m := newTestJWTManager(t)
	token, err := m.CreateAccessToken("user1", fixtureAppUserID, "org1", "jti-expiry")
	require.NoError(t, err)

	// when
	userInfo, _, err := m.ParseAccessToken(token)

	// then
	require.NoError(t, err)
	assert.False(t, userInfo.ExpiresAt.IsZero())
	assert.True(t, userInfo.ExpiresAt.After(time.Now()))
}
