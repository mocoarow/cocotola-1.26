package token

import (
	"errors"
	"sort"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// WhitelistEntry holds the minimal information needed for whitelist eviction.
type WhitelistEntry struct {
	ID        string
	CreatedAt time.Time
}

// Whitelist is an aggregate that manages a set of token IDs for a single user,
// enforcing the "keep at most maxSize tokens per user" invariant.
// It is used by SessionTokenWhitelist, RefreshTokenWhitelist, and AccessTokenWhitelist
// which share the same eviction logic.
type Whitelist struct {
	userID  domain.AppUserID
	entries []WhitelistEntry
	maxSize int
}

// NewWhitelist creates a new Whitelist with validated parameters.
func NewWhitelist(userID domain.AppUserID, entries []WhitelistEntry, maxSize int) (*Whitelist, error) {
	if userID.IsZero() {
		return nil, errors.New("whitelist user id is required")
	}
	if maxSize <= 0 {
		return nil, errors.New("whitelist max size must be positive")
	}
	copied := make([]WhitelistEntry, len(entries))
	copy(copied, entries)
	return &Whitelist{
		userID:  userID,
		entries: copied,
		maxSize: maxSize,
	}, nil
}

// UserID returns the user ID this whitelist belongs to.
func (w *Whitelist) UserID() domain.AppUserID {
	return w.userID
}

// Entries returns a defensive copy of the current entries.
func (w *Whitelist) Entries() []WhitelistEntry {
	copied := make([]WhitelistEntry, len(w.entries))
	copy(copied, w.entries)
	return copied
}

// ContainsToken returns true if the given token ID exists in the whitelist.
func (w *Whitelist) ContainsToken(tokenID string) bool {
	for _, e := range w.entries {
		if e.ID == tokenID {
			return true
		}
	}
	return false
}

// Remove removes entries with the given IDs from the whitelist.
func (w *Whitelist) Remove(tokenIDs []string) {
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
func (w *Whitelist) Add(entry WhitelistEntry) []string {
	w.entries = append(w.entries, entry)

	if len(w.entries) <= w.maxSize {
		return nil
	}

	sort.Slice(w.entries, func(i, j int) bool {
		return w.entries[i].CreatedAt.Before(w.entries[j].CreatedAt)
	})

	evictCount := len(w.entries) - w.maxSize
	evictedIDs := make([]string, evictCount)

	for i := range evictCount {
		evictedIDs[i] = w.entries[i].ID
	}

	w.entries = w.entries[evictCount:]

	return evictedIDs
}
