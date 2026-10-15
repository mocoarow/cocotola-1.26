package gateway_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

func jwksBody(t *testing.T, key *rsa.PrivateKey) []byte {
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
	return data
}

func serveJWKS(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()
	server, _ := serveSwitchableJWKS(t, key)
	return server
}

func serveSwitchableJWKS(t *testing.T, key *rsa.PrivateKey) (*httptest.Server, *atomic.Bool) {
	t.Helper()
	data := jwksBody(t, key)
	failing := &atomic.Bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if failing.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	return server, failing
}

func makeRSAJWT(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	return makeRSAJWTWithKID(t, key, testKeyID, claims)
}

func makeRSAJWTWithKID(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"sub":            "user-uuid-123",
		"email":          "test@example.com",
		"email_verified": true,
		"exp":            time.Now().Add(1 * time.Hour).Unix(),
	}
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

func Test_SupabaseVerifier_Verify_shouldNotIncludeEmailInError_whenEmailNotVerified(t *testing.T) {
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
	assert.NotContains(t, err.Error(), "test@example.com")
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

func Test_SupabaseVerifier_Verify_shouldReturnInternal_whenKIDIsUnknownAndJWKSRefreshFails(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	rotatedKey := generateTestKey(t)
	server, failing := serveSwitchableJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()
	failing.Store(true)

	tokenStr := makeRSAJWTWithKID(t, rotatedKey, "rotated-key-id", validClaims())

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrInternal)
}

func Test_SupabaseVerifier_Verify_shouldReturnUnauthenticated_whenKIDIsUnknownAndJWKSIsAvailable(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	unknownKey := generateTestKey(t)
	server := serveJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWTWithKID(t, unknownKey, "unknown-key-id", validClaims())

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnUnauthenticated_whenKIDIsUnknownAfterJWKSRecovers(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	unknownKey := generateTestKey(t)
	server, failing := serveSwitchableJWKS(t, key)
	failing.Store(true)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()
	failing.Store(false)

	tokenStr := makeRSAJWTWithKID(t, unknownKey, "unknown-key-id", validClaims())

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnSubAndEmail_whenKeyIsCachedDuringJWKSOutage(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	verifier := newVerifierDuringJWKSOutage(t, key)
	tokenStr := makeRSAJWT(t, key, validClaims())

	// when
	sub, _, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.NoError(t, err)
	assert.Equal(t, "user-uuid-123", sub)
}

func Test_SupabaseVerifier_Verify_shouldReturnUnauthenticated_whenEarlierRefreshWasCanceledByClient(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	unknownKey := generateTestKey(t)
	data := jwksBody(t, key)
	refreshStarted := make(chan struct{}, 1)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) > 1 {
			refreshStarted <- struct{}{}
			<-r.Context().Done()
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()

	tokenStr := makeRSAJWTWithKID(t, unknownKey, "unknown-key-id", validClaims())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-refreshStarted:
			cancel()
		case <-ctx.Done():
		}
	}()
	_, _, err = verifier.Verify(ctx, tokenStr)
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
	require.Equal(t, int32(2), requests.Load())

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func newVerifierDuringJWKSOutage(t *testing.T, key *rsa.PrivateKey) *gateway.SupabaseVerifier {
	t.Helper()
	server, failing := serveSwitchableJWKS(t, key)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	t.Cleanup(verifier.Close)
	failing.Store(true)
	rotatedTokenStr := makeRSAJWTWithKID(t, generateTestKey(t), "rotated-key-id", validClaims())
	_, _, err = verifier.Verify(context.Background(), rotatedTokenStr)
	require.ErrorIs(t, err, domain.ErrInternal)
	return verifier
}

func Test_SupabaseVerifier_Verify_shouldReturnUnauthenticated_whenTokenIsExpiredDuringJWKSOutage(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	verifier := newVerifierDuringJWKSOutage(t, key)
	claims := validClaims()
	claims["exp"] = time.Now().Add(-1 * time.Hour).Unix()
	tokenStr := makeRSAJWT(t, key, claims)

	// when
	_, _, err := verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnUnauthenticated_whenAlgIsUnknownDuringJWKSOutage(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	verifier := newVerifierDuringJWKSOutage(t, key)
	header, err := json.Marshal(map[string]any{"alg": "XYZ", "kid": testKeyID, "typ": "JWT"})
	require.NoError(t, err)
	payload, err := json.Marshal(validClaims())
	require.NoError(t, err)
	tokenStr := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + ".c2ln"

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_SupabaseVerifier_Verify_shouldReturnInternal_whenAlgMismatchesCachedKeyDuringJWKSOutage(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	verifier := newVerifierDuringJWKSOutage(t, key)
	token := jwt.NewWithClaims(jwt.SigningMethodRS384, validClaims())
	token.Header["kid"] = testKeyID
	tokenStr, err := token.SignedString(key)
	require.NoError(t, err)

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrInternal)
}

func Test_SupabaseVerifier_Verify_shouldReturnInternal_whenKIDIsUnknownAndJWKSReturnsInvalidBody(t *testing.T) {
	t.Parallel()

	// given
	key := generateTestKey(t)
	data := jwksBody(t, key)
	invalidBody := &atomic.Bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if invalidBody.Load() {
			_, _ = w.Write([]byte("<html>maintenance</html>"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	verifier, err := gateway.NewSupabaseVerifier(context.Background(), server.URL)
	require.NoError(t, err)
	defer verifier.Close()
	invalidBody.Store(true)

	tokenStr := makeRSAJWTWithKID(t, generateTestKey(t), "rotated-key-id", validClaims())

	// when
	_, _, err = verifier.Verify(context.Background(), tokenStr)

	// then
	require.ErrorIs(t, err, domain.ErrInternal)
}
