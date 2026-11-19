package gateway

import (
	"context"
	"time"
)

const organizationIDCacheMaxEntries = 1000

// NewCachedOrganizationResolver caches successful resolutions of resolver for ttl.
func NewCachedOrganizationResolver(resolver func(ctx context.Context, name string) (string, error), ttl time.Duration) func(ctx context.Context, name string) (string, error) {
	return newCachedOrganizationResolver(resolver, ttl, time.Now)
}

func newCachedOrganizationResolver(resolver func(ctx context.Context, name string) (string, error), ttl time.Duration, now func() time.Time) func(ctx context.Context, name string) (string, error) {
	if ttl <= 0 {
		return resolver
	}
	cache := newTTLCache[string](organizationIDCacheMaxEntries, now)

	return func(ctx context.Context, name string) (string, error) {
		if entry, ok := cache.get(name); ok {
			return entry.value, nil
		}
		id, err := resolver(ctx, name)
		if err != nil {
			return "", err
		}
		cache.set(name, id, now().Add(ttl))
		return id, nil
	}
}
