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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func (s *meAuthServer) fetcher(ttl time.Duration, clock *testClock) *gateway.AuthServiceMeFetcher {
	return gateway.NewAuthServiceMeFetcherWithClock(s.url, s.client, ttl, clock.now)
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

func Test_AuthServiceMeFetcher_FetchMe_shouldNotCallAuth_whenSameTokenIsWithinTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	_, err := fetcher.FetchMe(context.Background(), token)
	require.NoError(t, err)
	clock.advance(testMeCacheTTL - time.Second)

	// when
	_, err = fetcher.FetchMe(context.Background(), token)

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(1), server.calls.Load())
}

func Test_AuthServiceMeFetcher_FetchMe_shouldReturnCachedUser_whenSameTokenIsWithinTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	_, err := fetcher.FetchMe(context.Background(), token)
	require.NoError(t, err)

	// when
	me, err := fetcher.FetchMe(context.Background(), token)

	// then
	require.NoError(t, err)
	assert.Equal(t, "user-42", me.UserID)
}

func Test_AuthServiceMeFetcher_FetchMe_shouldCallAuthAgain_whenTTLElapsed(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	_, err := fetcher.FetchMe(context.Background(), token)
	require.NoError(t, err)
	clock.advance(testMeCacheTTL)

	// when
	_, err = fetcher.FetchMe(context.Background(), token)

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_AuthServiceMeFetcher_FetchMe_shouldCallAuthAgain_whenTokenExpiresBeforeTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(30*time.Second))
	_, err := fetcher.FetchMe(context.Background(), token)
	require.NoError(t, err)
	clock.advance(30 * time.Second)

	// when
	_, err = fetcher.FetchMe(context.Background(), token)

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_AuthServiceMeFetcher_FetchMe_shouldCallAuthEveryTime_whenTokenHasNoExp(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	_, err := fetcher.FetchMe(context.Background(), "Bearer opaque-token")
	require.NoError(t, err)

	// when
	_, err = fetcher.FetchMe(context.Background(), "Bearer opaque-token")

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_AuthServiceMeFetcher_FetchMe_shouldCallAuthEveryTime_whenAuthReturns401(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusUnauthorized)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	_, _ = fetcher.FetchMe(context.Background(), token)

	// when
	_, _ = fetcher.FetchMe(context.Background(), token)

	// then
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_AuthServiceMeFetcher_FetchMe_shouldCallAuthEveryTime_whenTTLIsZero(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(0, clock)
	token := bearerJWT(t, "user-42", clock.now().Add(time.Hour))
	_, err := fetcher.FetchMe(context.Background(), token)
	require.NoError(t, err)

	// when
	_, err = fetcher.FetchMe(context.Background(), token)

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_AuthServiceMeFetcher_FetchMe_shouldCallAuth_whenAnotherTokenIsCached(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	clock := newTestClock()
	fetcher := server.fetcher(testMeCacheTTL, clock)
	_, err := fetcher.FetchMe(context.Background(), bearerJWT(t, "user-42", clock.now().Add(time.Hour)))
	require.NoError(t, err)

	// when
	_, err = fetcher.FetchMe(context.Background(), bearerJWT(t, "user-43", clock.now().Add(time.Hour)))

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}
