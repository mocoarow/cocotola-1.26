package gateway_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

const testKeyID = "test-key-id"

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func serveJWKS(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": testKeyID,
				"use": "sig",
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
			},
		},
	}
	data, err := json.Marshal(jwks)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	return server
}

func makeRSAJWT(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func Test_SupabaseVerifier_Verify_shouldReturnSubAndEmail_whenTokenIsValid(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"sub":            "user-uuid-123",
		"email":          "test@example.com",
		"email_verified": true,
		"exp":            time.Now().Add(1 * time.Hour).Unix(),
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
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	})

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenKeyIsWrong(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	wrongKey := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, wrongKey, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenSubIsMissing(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenEmailIsMissing(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"sub": "user-uuid-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenEmailNotVerified(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrSupabaseEmailNotVerified)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenEmailVerifiedIsFalse(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"sub":            "user-uuid-123",
		"email":          "test@example.com",
		"email_verified": false,
		"exp":            time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrSupabaseEmailNotVerified)
}

func Test_SupabaseVerifier_Verify_shouldReturnSubAndEmail_whenUserMetadataEmailVerifiedIsTrue(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWT(t, key, jwt.MapClaims{
		"sub":   "user-uuid-123",
		"email": "test@example.com",
		"user_metadata": map[string]any{
			"email_verified": true,
		},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	// when
	sub, email, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.NoError(t, err)
	assert.Equal(t, "user-uuid-123", sub)
	assert.Equal(t, "test@example.com", email)
}

func Test_SupabaseVerifier_Verify_shouldReturnError_whenSigningMethodIsNotRSA(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

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
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}
