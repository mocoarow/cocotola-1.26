package gateway

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

const (
	authMeCacheMaxEntries = 10000
	jwtSegmentCount       = 3
)

type authMeCache struct {
	ttl     time.Duration
	now     func() time.Time
	entries *ttlCache[getMeResponse]
}

func newAuthMeCache(ttl time.Duration, now func() time.Time) *authMeCache {
	return &authMeCache{
		ttl:     ttl,
		now:     now,
		entries: newTTLCache[getMeResponse](authMeCacheMaxEntries, now),
	}
}

func (c *authMeCache) get(authorization string) (getMeResponse, bool) {
	var me getMeResponse
	if c.ttl <= 0 {
		return me, false
	}
	entry, ok := c.entries.get(authMeCacheKey(authorization))
	if !ok {
		return me, false
	}
	return entry.value, true
}

// set must be called only after cocotola-auth accepted authorization:
// the JWT exp is read without verifying the signature.
func (c *authMeCache) set(authorization string, me getMeResponse) {
	if c.ttl <= 0 {
		return
	}
	exp, ok := bearerTokenExpiry(authorization)
	if !ok {
		return
	}
	expiresAt := c.now().Add(c.ttl)
	if exp.Before(expiresAt) {
		expiresAt = exp
	}
	c.entries.set(authMeCacheKey(authorization), me, expiresAt)
}

func authMeCacheKey(authorization string) string {
	sum := sha256.Sum256([]byte(authorization))
	return hex.EncodeToString(sum[:])
}

func bearerTokenExpiry(authorization string) (time.Time, bool) {
	var zero time.Time
	token, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok {
		return zero, false
	}
	segments := strings.Split(token, ".")
	if len(segments) != jwtSegmentCount {
		return zero, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return zero, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp <= 0 {
		return zero, false
	}
	return time.Unix(claims.Exp, 0), true
}
