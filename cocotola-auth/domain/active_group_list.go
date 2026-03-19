package domain

import "errors"

// ActiveGroupList is an aggregate that manages the set of active group IDs for an organization,
// enforcing the "at most maxActiveGroups" invariant.
type ActiveGroupList struct{ intIDSet }

// NewActiveGroupList creates a validated ActiveGroupList.
func NewActiveGroupList(organizationID int, entries []int) (*ActiveGroupList, error) {
	if organizationID <= 0 {
		return nil, errors.New("active group list organization id must be positive")
	}
	return &ActiveGroupList{newIntIDSet(organizationID, entries)}, nil
}

// OrganizationID returns the organization ID.
func (l *ActiveGroupList) OrganizationID() int { return l.getOwnerID() }

// Entries returns a copy of the current group IDs as a slice.
func (l *ActiveGroupList) Entries() []int { return l.ids() }

// Add adds a group ID to the list. Returns ErrActiveGroupLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the group ID already exists.
func (l *ActiveGroupList) Add(groupID int, maxActiveGroups int) error {
	return l.addWithLimit(groupID, maxActiveGroups, ErrActiveGroupLimitReached)
}
