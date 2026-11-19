package gateway

import (
	"sync"
	"time"
)

type ttlCacheEntry[V any] struct {
	value     V
	expiresAt time.Time
}

type ttlCache[V any] struct {
	mu         sync.Mutex
	entries    map[string]ttlCacheEntry[V]
	maxEntries int
	now        func() time.Time
}

func newTTLCache[V any](maxEntries int, now func() time.Time) *ttlCache[V] {
	return &ttlCache[V]{
		mu:         sync.Mutex{},
		entries:    make(map[string]ttlCacheEntry[V]),
		maxEntries: maxEntries,
		now:        now,
	}
}

func (c *ttlCache[V]) get(key string) (ttlCacheEntry[V], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var miss ttlCacheEntry[V]
	entry, ok := c.entries[key]
	if !ok {
		return miss, false
	}
	if !c.now().Before(entry.expiresAt) {
		delete(c.entries, key)
		return miss, false
	}
	return entry, true
}

func (c *ttlCache[V]) set(key string, value V, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if !now.Before(expiresAt) {
		return
	}
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.maxEntries {
		c.sweepExpired(now)
		if len(c.entries) >= c.maxEntries {
			return
		}
	}
	c.entries[key] = ttlCacheEntry[V]{value: value, expiresAt: expiresAt}
}

func (c *ttlCache[V]) sweepExpired(now time.Time) {
	for key, entry := range c.entries {
		if !now.Before(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}
