package domain

import (
	"fmt"
	"sort"
)

// OwnedWorkbookList is an aggregate that manages the set of workbook IDs
// owned by a user, enforcing the "at most maxWorkbooks" invariant.
type OwnedWorkbookList struct {
	ownerID string
	version int
	entries map[string]struct{}
}

// NewOwnedWorkbookList creates a validated OwnedWorkbookList.
func NewOwnedWorkbookList(ownerID string, workbookIDs []string) (*OwnedWorkbookList, error) {
	if ownerID == "" {
		return nil, fmt.Errorf("owned workbook list owner id is required: %w", ErrInvalidArgument)
	}
	m := make(map[string]struct{}, len(workbookIDs))
	for _, id := range workbookIDs {
		if id == "" {
			return nil, fmt.Errorf("workbook id is required: %w", ErrInvalidArgument)
		}
		m[id] = struct{}{}
	}
	return &OwnedWorkbookList{ownerID: ownerID, entries: m}, nil
}

// Add adds a workbook ID to the list. Returns ErrOwnedWorkbookLimitReached if
// the list is at capacity, or ErrDuplicateOwnedWorkbook if the workbook ID
// already exists.
func (l *OwnedWorkbookList) Add(workbookID string, maxWorkbooks int) error {
	if workbookID == "" {
		return fmt.Errorf("workbook id is required: %w", ErrInvalidArgument)
	}
	if maxWorkbooks <= 0 {
		return fmt.Errorf("max workbooks must be positive, got %d: %w", maxWorkbooks, ErrInvalidArgument)
	}
	if _, ok := l.entries[workbookID]; ok {
		return ErrDuplicateOwnedWorkbook
	}
	if len(l.entries) >= maxWorkbooks {
		return ErrOwnedWorkbookLimitReached
	}
	l.entries[workbookID] = struct{}{}
	return nil
}

// Remove removes a workbook ID from the list.
func (l *OwnedWorkbookList) Remove(workbookID string) {
	delete(l.entries, workbookID)
}

// OwnerID returns the owner user ID.
func (l *OwnedWorkbookList) OwnerID() string { return l.ownerID }

// Version returns the persisted version (0 = new, not yet saved).
func (l *OwnedWorkbookList) Version() int { return l.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
func (l *OwnedWorkbookList) SetVersion(version int) {
	l.version = version
}

// Entries returns a sorted copy of the current workbook IDs as a slice.
func (l *OwnedWorkbookList) Entries() []string {
	result := make([]string, 0, len(l.entries))
	for id := range l.entries {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

// Size returns the number of entries in the set.
func (l *OwnedWorkbookList) Size() int { return len(l.entries) }

// Contains returns true if the given workbook ID exists in the set.
func (l *OwnedWorkbookList) Contains(workbookID string) bool {
	_, ok := l.entries[workbookID]
	return ok
}
