package domain

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// ActiveGroupList is an aggregate that manages the set of active group IDs for an organization,
// enforcing the "at most maxActiveGroups" invariant.
type ActiveGroupList struct {
	activeListBase[GroupID]
}

// NewActiveGroupList creates a validated ActiveGroupList.
func NewActiveGroupList(organizationID OrganizationID, entries []GroupID) (*ActiveGroupList, error) {
	if organizationID.IsZero() {
		return nil, errors.New("active group list organization id must not be zero")
	}
	return &ActiveGroupList{activeListBase[GroupID]{idset.New[OrganizationID, GroupID](organizationID, entries)}}, nil
}

// Add adds a group ID to the list. Returns ErrActiveGroupLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the group ID already exists.
func (l *ActiveGroupList) Add(groupID GroupID, maxActiveGroups int) error {
	if err := l.set.AddWithLimit(groupID, maxActiveGroups, ErrActiveGroupLimitReached, ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add group to active list: %w", err)
	}
	return nil
}
