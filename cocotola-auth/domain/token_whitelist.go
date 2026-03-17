package domain

import (
	"sort"
	"time"
)

// WhitelistEntry holds the minimal information needed for whitelist eviction.
type WhitelistEntry struct {
	ID        string
	CreatedAt time.Time
}

// TokenWhitelist is an aggregate that manages a set of token IDs for a single user,
// enforcing the "keep at most maxSize tokens per user" invariant.
// It is used by SessionTokenWhitelist, RefreshTokenWhitelist, and AccessTokenWhitelist
// which share the same eviction logic.
type TokenWhitelist struct {
	userID  int
	entries []WhitelistEntry
	maxSize int
}

// NewTokenWhitelist creates a new TokenWhitelist.
func NewTokenWhitelist(userID int, entries []WhitelistEntry, maxSize int) *TokenWhitelist {
	copied := make([]WhitelistEntry, len(entries))
	copy(copied, entries)
	return &TokenWhitelist{
		userID:  userID,
		entries: copied,
		maxSize: maxSize,
	}
}

// UserID returns the user ID this whitelist belongs to.
func (w *TokenWhitelist) UserID() int {
	return w.userID
}

// Entries returns a defensive copy of the current entries.
func (w *TokenWhitelist) Entries() []WhitelistEntry {
	copied := make([]WhitelistEntry, len(w.entries))
	copy(copied, w.entries)
	return copied
}

// ContainsToken returns true if the given token ID exists in the whitelist.
func (w *TokenWhitelist) ContainsToken(tokenID string) bool {
	for _, e := range w.entries {
		if e.ID == tokenID {
			return true
		}
	}
	return false
}

// Remove removes entries with the given IDs from the whitelist.
func (w *TokenWhitelist) Remove(tokenIDs []string) {
	if len(tokenIDs) == 0 {
		return
	}
	removeSet := make(map[string]struct{}, len(tokenIDs))
	for _, id := range tokenIDs {
		removeSet[id] = struct{}{}
	}
	filtered := make([]WhitelistEntry, 0, len(w.entries))
	for _, e := range w.entries {
		if _, ok := removeSet[e.ID]; !ok {
			filtered = append(filtered, e)
		}
	}
	w.entries = filtered
}

// Add appends a new entry and returns the IDs of entries that should be evicted
// to keep the whitelist within maxSize. Evicts oldest by CreatedAt first.
func (w *TokenWhitelist) Add(entry WhitelistEntry) []string {
	w.entries = append(w.entries, entry)

	if len(w.entries) <= w.maxSize {
		return nil
	}

	sort.Slice(w.entries, func(i, j int) bool {
		return w.entries[i].CreatedAt.Before(w.entries[j].CreatedAt)
	})

	evictCount := len(w.entries) - w.maxSize
	evictedIDs := make([]string, evictCount)
	for i := 0; i < evictCount; i++ {
		evictedIDs[i] = w.entries[i].ID
	}

	w.entries = w.entries[evictCount:]

	return evictedIDs
}
