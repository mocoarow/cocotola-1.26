package token

import (
	"fmt"
	"slices"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// WhitelistEntry holds the minimal information needed for whitelist eviction.
type WhitelistEntry struct {
	ID        string
	CreatedAt time.Time
}

// Whitelist is an aggregate holding the token IDs that are still valid for a single user.
type Whitelist struct {
	userID  domain.AppUserID
	entries []WhitelistEntry
	version int
}

// NewWhitelist creates a new Whitelist with validated parameters.
func NewWhitelist(userID domain.AppUserID, entries []WhitelistEntry) (*Whitelist, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("whitelist user id is required: %w", domain.ErrInvalidArgument)
	}
	return &Whitelist{
		userID:  userID,
		entries: slices.Clone(entries),
		version: 0,
	}, nil
}

// UserID returns the user ID this whitelist belongs to.
func (w *Whitelist) UserID() domain.AppUserID {
	return w.userID
}

// Entries returns a defensive copy of the current entries.
func (w *Whitelist) Entries() []WhitelistEntry {
	return slices.Clone(w.entries)
}

// Version returns the persisted version (0 = not yet saved).
func (w *Whitelist) Version() int { return w.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
func (w *Whitelist) SetVersion(version int) { w.version = version }

// ContainsToken returns true if the given token ID exists in the whitelist.
func (w *Whitelist) ContainsToken(tokenID string) bool {
	return slices.ContainsFunc(w.entries, func(e WhitelistEntry) bool { return e.ID == tokenID })
}

// Remove removes entries with the given IDs from the whitelist.
func (w *Whitelist) Remove(tokenIDs []string) {
	w.entries = slices.DeleteFunc(slices.Clone(w.entries), func(e WhitelistEntry) bool { return slices.Contains(tokenIDs, e.ID) })
}

// Add appends entry and evicts the oldest entries by CreatedAt beyond maxSize.
func (w *Whitelist) Add(entry WhitelistEntry, maxSize int) error {
	if maxSize <= 0 {
		return fmt.Errorf("whitelist max size must be positive: %w", domain.ErrInvalidArgument)
	}
	if w.ContainsToken(entry.ID) {
		return fmt.Errorf("token %s is already whitelisted: %w", entry.ID, domain.ErrDuplicateEntry)
	}
	entries := append(slices.Clone(w.entries), entry)
	slices.SortStableFunc(entries, func(a, b WhitelistEntry) int { return a.CreatedAt.Compare(b.CreatedAt) })
	if len(entries) > maxSize {
		entries = entries[len(entries)-maxSize:]
	}
	w.entries = entries
	return nil
}
