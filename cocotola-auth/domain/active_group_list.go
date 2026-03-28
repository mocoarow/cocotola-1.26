package domain

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// ActiveGroupList is an aggregate that manages the set of active group IDs for an organization,
// enforcing the "at most maxActiveGroups" invariant.
type ActiveGroupList struct{ activeListBase }

// NewActiveGroupList creates a validated ActiveGroupList.
func NewActiveGroupList(organizationID int, entries []int) (*ActiveGroupList, error) {
	if organizationID <= 0 {
		return nil, errors.New("active group list organization id must be positive")
	}
	return &ActiveGroupList{activeListBase{idset.New(organizationID, entries)}}, nil
}

// Add adds a group ID to the list. Returns ErrActiveGroupLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the group ID already exists.
func (l *ActiveGroupList) Add(groupID int, maxActiveGroups int) error {
	if err := l.set.AddWithLimit(groupID, maxActiveGroups, ErrActiveGroupLimitReached, ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add group to active list: %w", err)
	}
	return nil
}
