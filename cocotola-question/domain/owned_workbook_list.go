package domain

import (
	"fmt"
	"slices"
	"sort"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"
)

// OwnedWorkbookList is an aggregate that manages the set of workbook IDs
// owned by a user, enforcing the "at most maxWorkbooks" invariant.
type OwnedWorkbookList struct {
	set     idset.Set[string, string]
	version int
}

// NewOwnedWorkbookList creates a validated OwnedWorkbookList.
func NewOwnedWorkbookList(ownerID string, workbookIDs []string) (*OwnedWorkbookList, error) {
	if ownerID == "" {
		return nil, fmt.Errorf("owned workbook list owner id is required: %w", ErrInvalidArgument)
	}
	if slices.Contains(workbookIDs, "") {
		return nil, fmt.Errorf("workbook id is required: %w", ErrInvalidArgument)
	}
	return &OwnedWorkbookList{set: idset.New(ownerID, workbookIDs), version: 0}, nil
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
	if err := l.set.AddWithLimit(workbookID, maxWorkbooks, ErrOwnedWorkbookLimitReached, ErrDuplicateOwnedWorkbook); err != nil {
		return fmt.Errorf("add workbook to owned list: %w", err)
	}
	return nil
}

// Remove removes a workbook ID from the list.
func (l *OwnedWorkbookList) Remove(workbookID string) { l.set.Remove(workbookID) }

// OwnerID returns the owner user ID.
func (l *OwnedWorkbookList) OwnerID() string { return l.set.OwnerID }

// Version returns the persisted version (0 = new, not yet saved).
func (l *OwnedWorkbookList) Version() int { return l.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
func (l *OwnedWorkbookList) SetVersion(version int) {
	l.version = version
}

// Entries returns a sorted copy of the current workbook IDs as a slice.
func (l *OwnedWorkbookList) Entries() []string {
	result := l.set.IDs()
	sort.Strings(result)
	return result
}

// Size returns the number of entries in the set.
func (l *OwnedWorkbookList) Size() int { return l.set.Size() }

// Contains returns true if the given workbook ID exists in the set.
func (l *OwnedWorkbookList) Contains(workbookID string) bool { return l.set.Contains(workbookID) }
