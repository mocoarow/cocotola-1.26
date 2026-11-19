package gateway_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const testMeCacheTTL = 90 * time.Second

type meAuthServer struct {
	calls  atomic.Int64
	url    string
	client *http.Client
}

func newMeAuthServer(t *testing.T, status int) *meAuthServer {
	t.Helper()
	s := &meAuthServer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.calls.Add(1)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"userId":"user-42","loginId":"login-42","organizationName":"org-42","language":"ja"}`))
	}))
	t.Cleanup(server.Close)
	s.url = server.URL
	s.client = server.Client()
	return s
}

func (s *meAuthServer) router(ttl time.Duration, clock *testClock) *gin.Engine {
	r := gin.New()
	r.Use(gateway.NewAuthMiddlewareWithClock(s.url, s.client, ttl, clock.now))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString(controller.ContextFieldUserID{}))
	})
	return r
}

func bearerJWT(t *testing.T, subject string, exp time.Time) string {
	t.Helper()
	encode := func(v any) string {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	header := encode(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload := encode(map[string]any{"sub": subject, "exp": exp.Unix()})
	return "Bearer " + header + "." + payload + ".signature"
}

func serveWithAuthorization(t *testing.T, r *gin.Engine, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", authorization)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Test_NewAuthMiddleware_shouldNotCallAuth_whenSameTokenIsWithinTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	serveWithAuthorization(t, r, token)
	clock.advance(testMeCacheTTL - time.Second)

	// when
	serveWithAuthorization(t, r, token)

	// then
	assert.Equal(t, int64(1), server.calls.Load())
}

func Test_NewAuthMiddleware_shouldSetUserIDFromCache_whenSameTokenIsWithinTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	serveWithAuthorization(t, r, token)

	// when
	w := serveWithAuthorization(t, r, token)

	// then
	assert.Equal(t, "user-42", w.Body.String())
}

func Test_NewAuthMiddleware_shouldCallAuthAgain_whenTTLElapsed(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	serveWithAuthorization(t, r, token)
	clock.advance(testMeCacheTTL)

	// when
	serveWithAuthorization(t, r, token)

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewAuthMiddleware_shouldCallAuthAgain_whenTokenExpiresBeforeTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(30*time.Second))
	serveWithAuthorization(t, r, token)
	clock.advance(30 * time.Second)

	// when
	serveWithAuthorization(t, r, token)

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewAuthMiddleware_shouldCallAuthEveryTime_whenTokenHasNoExp(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	serveWithAuthorization(t, r, "Bearer opaque-token")

	// when
	serveWithAuthorization(t, r, "Bearer opaque-token")

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewAuthMiddleware_shouldCallAuthEveryTime_whenAuthReturns401(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusUnauthorized)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	serveWithAuthorization(t, r, token)

	// when
	serveWithAuthorization(t, r, token)

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewAuthMiddleware_shouldCallAuthEveryTime_whenTTLIsZero(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(0, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	serveWithAuthorization(t, r, token)

	// when
	serveWithAuthorization(t, r, token)

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewAuthMiddleware_shouldCallAuth_whenAnotherTokenIsCached(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	r := server.router(testMeCacheTTL, clock)
	serveWithAuthorization(t, r, bearerJWT(t, "user-42", clock.now().Add(time.Hour)))

	// when
	serveWithAuthorization(t, r, bearerJWT(t, "user-43", clock.now().Add(time.Hour)))

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}
