package gateway_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const testOrgCacheTTL = 60 * time.Minute

type testClock struct {
	nanos atomic.Int64
}

func newTestClock() *testClock {
	c := &testClock{}
	c.nanos.Store(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).UnixNano())
	return c
}

func (c *testClock) now() time.Time {
	return time.Unix(0, c.nanos.Load()).UTC()
}

func (c *testClock) advance(d time.Duration) {
	c.nanos.Add(int64(d))
}

// orgAuthServer answers the organization lookup with "id-<name>" and fails
// the first failFirst requests with 500.
type orgAuthServer struct {
	calls     atomic.Int64
	failFirst int64
	url       string
	client    *http.Client
}

func newOrgAuthServer(t *testing.T, failFirst int64) *orgAuthServer {
	t.Helper()
	s := &orgAuthServer{failFirst: failFirst}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := s.calls.Add(1)
		if n <= s.failFirst {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "id-" + r.URL.Query().Get("name")})
	}))
	t.Cleanup(server.Close)
	s.url = server.URL
	s.client = server.Client()
	return s
}

func (s *orgAuthServer) resolver(ttl time.Duration, clock *testClock) func(ctx context.Context, name string) (string, error) {
	return gateway.NewCachedOrganizationResolverWithClock(
		gateway.AuthServiceOrganizationResolver(s.url, testAPIKey, s.client), ttl, clock.now)
}

func Test_NewCachedOrganizationResolver_shouldNotCallAuth_whenNameWasResolvedWithinTTL(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 0)
	clock := newTestClock()
	resolve := server.resolver(testOrgCacheTTL, clock)
	_, err := resolve(context.Background(), "org-a")
	require.NoError(t, err)
	clock.advance(testOrgCacheTTL - time.Second)

	// when
	id, err := resolve(context.Background(), "org-a")

	// then
	require.NoError(t, err)
	assert.Equal(t, "id-org-a", id)
	assert.Equal(t, int64(1), server.calls.Load())
}

func Test_NewCachedOrganizationResolver_shouldCallAuthAgain_whenTTLElapsed(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 0)
	clock := newTestClock()
	resolve := server.resolver(testOrgCacheTTL, clock)
	_, err := resolve(context.Background(), "org-a")
	require.NoError(t, err)
	clock.advance(testOrgCacheTTL)

	// when
	_, err = resolve(context.Background(), "org-a")

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewCachedOrganizationResolver_shouldCallAuthAgain_whenPreviousResolutionFailed(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 1)
	clock := newTestClock()
	resolve := server.resolver(testOrgCacheTTL, clock)
	firstID, _ := resolve(context.Background(), "org-a")
	require.Empty(t, firstID)

	// when
	_, err := resolve(context.Background(), "org-a")

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewCachedOrganizationResolver_shouldCallAuthEveryTime_whenTTLIsZero(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 0)
	clock := newTestClock()
	resolve := server.resolver(0, clock)
	_, err := resolve(context.Background(), "org-a")
	require.NoError(t, err)

	// when
	_, err = resolve(context.Background(), "org-a")

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(2), server.calls.Load())
}

func Test_NewCachedOrganizationResolver_shouldReturnOwnID_whenAnotherNameIsCached(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 0)
	clock := newTestClock()
	resolve := server.resolver(testOrgCacheTTL, clock)
	_, err := resolve(context.Background(), "org-a")
	require.NoError(t, err)

	// when
	id, err := resolve(context.Background(), "org-b")

	// then
	require.NoError(t, err)
	assert.Equal(t, "id-org-b", id)
}

func fillOrganizationCache(t *testing.T, resolve func(ctx context.Context, name string) (string, error)) {
	t.Helper()
	for i := range gateway.OrganizationIDCacheMaxEntries {
		_, err := resolve(context.Background(), fmt.Sprintf("filler-%d", i))
		require.NoError(t, err)
	}
}

func Test_NewCachedOrganizationResolver_shouldNotCacheNewName_whenCacheIsFullOfUnexpiredEntries(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 0)
	clock := newTestClock()
	resolve := server.resolver(testOrgCacheTTL, clock)
	fillOrganizationCache(t, resolve)
	_, err := resolve(context.Background(), "org-a")
	require.NoError(t, err)
	callsBefore := server.calls.Load()

	// when
	_, err = resolve(context.Background(), "org-a")

	// then
	require.NoError(t, err)
	assert.Equal(t, callsBefore+1, server.calls.Load())
}

func Test_NewCachedOrganizationResolver_shouldCacheNewName_whenFullCacheHasExpiredEntries(t *testing.T) {
	t.Parallel()

	// given
	server := newOrgAuthServer(t, 0)
	clock := newTestClock()
	resolve := server.resolver(testOrgCacheTTL, clock)
	fillOrganizationCache(t, resolve)
	clock.advance(testOrgCacheTTL)
	_, err := resolve(context.Background(), "org-a")
	require.NoError(t, err)
	callsBefore := server.calls.Load()

	// when
	_, err = resolve(context.Background(), "org-a")

	// then
	require.NoError(t, err)
	assert.Equal(t, callsBefore, server.calls.Load())
}
