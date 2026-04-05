package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func makeSupabaseJWT(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func Test_SupabaseVerifier_Verify_shouldReturnSubAndEmail_whenTokenIsValid(t *testing.T) {
	t.Parallel()

	// given
	secret := "test-secret-must-be-at-least-32-chars"
	verifier := gateway.NewSupabaseVerifier(secret)
	tokenStr := makeSupabaseJWT(t, secret, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	sub, email, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.NoError(t, err)
	assert.Equal(t, "user-uuid-123", sub)
	assert.Equal(t, "test@example.com", email)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenTokenIsExpired(t *testing.T) {
	t.Parallel()

	// given
	secret := "test-secret-must-be-at-least-32-chars"
	verifier := gateway.NewSupabaseVerifier(secret)
	tokenStr := makeSupabaseJWT(t, secret, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	})

	// when
	_, _, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse supabase token")
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenSecretIsWrong(t *testing.T) {
	t.Parallel()

	// given
	verifier := gateway.NewSupabaseVerifier("correct-secret-at-least-32-chars-long")
	tokenStr := makeSupabaseJWT(t, "wrong-secret-at-least-32-chars-long!!", jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse supabase token")
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenSubIsMissing(t *testing.T) {
	t.Parallel()

	// given
	secret := "test-secret-must-be-at-least-32-chars"
	verifier := gateway.NewSupabaseVerifier(secret)
	tokenStr := makeSupabaseJWT(t, secret, jwt.MapClaims{
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing sub")
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenEmailIsMissing(t *testing.T) {
	t.Parallel()

	// given
	secret := "test-secret-must-be-at-least-32-chars"
	verifier := gateway.NewSupabaseVerifier(secret)
	tokenStr := makeSupabaseJWT(t, secret, jwt.MapClaims{
		"sub": "user-uuid-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing email")
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenSigningMethodIsNotHMAC(t *testing.T) {
	t.Parallel()

	// given
	secret := "test-secret-must-be-at-least-32-chars"
	verifier := gateway.NewSupabaseVerifier(secret)
	// Use "none" algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse supabase token")
}
