package gateway_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

const (
	cachedTokenKey = "cache-key"
	cachedTokenTTL = time.Hour
)

func newCachedSessionToken(expiresAt time.Time) *domaintoken.SessionToken {
	return domaintoken.ReconstructSessionToken("st-1", fixtureAppUserID, "user1", "org1", domain.TokenHash("hash"), expiresAt.Add(-cachedTokenTTL), expiresAt, nil)
}

func newCachedAccessToken(expiresAt time.Time) *domaintoken.AccessToken {
	return domaintoken.ReconstructAccessToken("at-1", "rt-1", fixtureAppUserID, "user1", "org1", expiresAt.Add(-cachedTokenTTL), expiresAt, nil)
}

func Test_TokenCache_GetSessionToken_shouldReturnToken_whenCachedAndNotExpired(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()
	token := newCachedSessionToken(time.Now().Add(cachedTokenTTL))
	cache.SetSessionToken(cachedTokenKey, token)

	// when
	got, _ := cache.GetSessionToken(cachedTokenKey)

	// then
	assert.Same(t, token, got)
}

func Test_TokenCache_GetSessionToken_shouldReturnFalse_whenNotCached(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()

	// when
	_, ok := cache.GetSessionToken(cachedTokenKey)

	// then
	assert.False(t, ok)
}

func Test_TokenCache_GetSessionToken_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()
	cache.SetSessionToken(cachedTokenKey, newCachedSessionToken(time.Now().Add(-cachedTokenTTL)))

	// when
	_, ok := cache.GetSessionToken(cachedTokenKey)

	// then
	assert.False(t, ok)
}

func Test_TokenCache_DeleteSessionToken_shouldRemoveToken_whenCached(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()
	cache.SetSessionToken(cachedTokenKey, newCachedSessionToken(time.Now().Add(cachedTokenTTL)))

	// when
	cache.DeleteSessionToken(cachedTokenKey)

	// then
	_, ok := cache.GetSessionToken(cachedTokenKey)
	assert.False(t, ok)
}

func Test_TokenCache_GetAccessToken_shouldReturnToken_whenCachedAndNotExpired(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()
	token := newCachedAccessToken(time.Now().Add(cachedTokenTTL))
	cache.SetAccessToken(cachedTokenKey, token)

	// when
	got, _ := cache.GetAccessToken(cachedTokenKey)

	// then
	assert.Same(t, token, got)
}

func Test_TokenCache_GetAccessToken_shouldReturnFalse_whenNotCached(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()

	// when
	_, ok := cache.GetAccessToken(cachedTokenKey)

	// then
	assert.False(t, ok)
}

func Test_TokenCache_GetAccessToken_shouldReturnFalse_whenExpired(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()
	cache.SetAccessToken(cachedTokenKey, newCachedAccessToken(time.Now().Add(-cachedTokenTTL)))

	// when
	_, ok := cache.GetAccessToken(cachedTokenKey)

	// then
	assert.False(t, ok)
}

func Test_TokenCache_DeleteAccessToken_shouldRemoveToken_whenCached(t *testing.T) {
	t.Parallel()

	// given
	cache := gateway.NewTokenCache()
	cache.SetAccessToken(cachedTokenKey, newCachedAccessToken(time.Now().Add(cachedTokenTTL)))

	// when
	cache.DeleteAccessToken(cachedTokenKey)

	// then
	_, ok := cache.GetAccessToken(cachedTokenKey)
	assert.False(t, ok)
}

// Get judges expiry by the real clock, so these tests use tokens that are valid now
// but expired at the time passed to CleanExpired to observe the removal.

func Test_TokenCache_CleanExpired_shouldRemoveSessionToken_whenExpiredAtGivenTime(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now()
	cache := gateway.NewTokenCache()
	cache.SetSessionToken(cachedTokenKey, newCachedSessionToken(now.Add(cachedTokenTTL)))

	// when
	cache.CleanExpired(now.Add(2 * cachedTokenTTL))

	// then
	_, ok := cache.GetSessionToken(cachedTokenKey)
	assert.False(t, ok)
}

func Test_TokenCache_CleanExpired_shouldKeepSessionToken_whenNotExpiredAtGivenTime(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now()
	cache := gateway.NewTokenCache()
	cache.SetSessionToken(cachedTokenKey, newCachedSessionToken(now.Add(3*cachedTokenTTL)))

	// when
	cache.CleanExpired(now.Add(2 * cachedTokenTTL))

	// then
	_, ok := cache.GetSessionToken(cachedTokenKey)
	assert.True(t, ok)
}

func Test_TokenCache_CleanExpired_shouldRemoveAccessToken_whenExpiredAtGivenTime(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now()
	cache := gateway.NewTokenCache()
	cache.SetAccessToken(cachedTokenKey, newCachedAccessToken(now.Add(cachedTokenTTL)))

	// when
	cache.CleanExpired(now.Add(2 * cachedTokenTTL))

	// then
	_, ok := cache.GetAccessToken(cachedTokenKey)
	assert.False(t, ok)
}

func Test_TokenCache_CleanExpired_shouldKeepAccessToken_whenNotExpiredAtGivenTime(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now()
	cache := gateway.NewTokenCache()
	cache.SetAccessToken(cachedTokenKey, newCachedAccessToken(now.Add(3*cachedTokenTTL)))

	// when
	cache.CleanExpired(now.Add(2 * cachedTokenTTL))

	// then
	_, ok := cache.GetAccessToken(cachedTokenKey)
	assert.True(t, ok)
}
