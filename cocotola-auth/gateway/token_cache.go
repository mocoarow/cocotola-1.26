package gateway

import (
	"sync"
	"time"

	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

// TokenCache provides in-memory caching for session tokens and access tokens.
// Expired tokens are lazily evicted on Get operations.
type TokenCache struct {
	sessionTokens sync.Map
	accessTokens  sync.Map
}

// NewTokenCache returns a new TokenCache.
func NewTokenCache() *TokenCache {
	return &TokenCache{
		sessionTokens: sync.Map{},
		accessTokens:  sync.Map{},
	}
}

// SetSessionToken caches a session token by its hash.
func (c *TokenCache) SetSessionToken(hash string, token *domaintoken.SessionToken) {
	c.sessionTokens.Store(hash, token)
}

// GetSessionToken retrieves a cached session token by hash.
// Returns (nil, false) if the token is not found or has expired.
func (c *TokenCache) GetSessionToken(hash string) (*domaintoken.SessionToken, bool) {
	v, ok := c.sessionTokens.Load(hash)
	if !ok {
		return nil, false
	}
	token, ok := v.(*domaintoken.SessionToken)
	if !ok {
		c.sessionTokens.Delete(hash)
		return nil, false
	}
	if token.IsExpired(time.Now()) {
		c.sessionTokens.Delete(hash)
		return nil, false
	}
	return token, true
}

// DeleteSessionToken removes a session token from the cache.
func (c *TokenCache) DeleteSessionToken(hash string) {
	c.sessionTokens.Delete(hash)
}

// SetAccessToken caches an access token by its JTI.
func (c *TokenCache) SetAccessToken(jti string, token *domaintoken.AccessToken) {
	c.accessTokens.Store(jti, token)
}

// GetAccessToken retrieves a cached access token by JTI.
// Returns (nil, false) if the token is not found or has expired.
func (c *TokenCache) GetAccessToken(jti string) (*domaintoken.AccessToken, bool) {
	v, ok := c.accessTokens.Load(jti)
	if !ok {
		return nil, false
	}
	token, ok := v.(*domaintoken.AccessToken)
	if !ok {
		c.accessTokens.Delete(jti)
		return nil, false
	}
	if token.IsExpired(time.Now()) {
		c.accessTokens.Delete(jti)
		return nil, false
	}
	return token, true
}

// DeleteAccessToken removes an access token from the cache.
func (c *TokenCache) DeleteAccessToken(jti string) {
	c.accessTokens.Delete(jti)
}

// CleanExpired removes all expired tokens from both caches.
func (c *TokenCache) CleanExpired(now time.Time) {
	c.sessionTokens.Range(func(key, value any) bool {
		if token, ok := value.(*domaintoken.SessionToken); ok && token.IsExpired(now) {
			c.sessionTokens.Delete(key)
		}
		return true
	})
	c.accessTokens.Range(func(key, value any) bool {
		if token, ok := value.(*domaintoken.AccessToken); ok && token.IsExpired(now) {
			c.accessTokens.Delete(key)
		}
		return true
	})
}
