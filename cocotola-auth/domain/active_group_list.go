package domain

import "errors"

// ActiveGroupList is an aggregate that manages the set of active group IDs for an organization,
// enforcing the "at most maxActiveGroups" invariant.
type ActiveGroupList struct {
	organizationID int
	entries        map[int]struct{}
}

// NewActiveGroupList creates a validated ActiveGroupList.
func NewActiveGroupList(organizationID int, entries []int) (*ActiveGroupList, error) {
	if organizationID <= 0 {
		return nil, errors.New("active group list organization id must be positive")
	}
	m := make(map[int]struct{}, len(entries))
	for _, id := range entries {
		m[id] = struct{}{}
	}
	return &ActiveGroupList{
		organizationID: organizationID,
		entries:        m,
	}, nil
}

// OrganizationID returns the organization ID.
func (l *ActiveGroupList) OrganizationID() int { return l.organizationID }

// Entries returns a copy of the current group IDs as a slice.
func (l *ActiveGroupList) Entries() []int {
	result := make([]int, 0, len(l.entries))
	for id := range l.entries {
		result = append(result, id)
	}
	return result
}

// Size returns the number of active groups.
func (l *ActiveGroupList) Size() int { return len(l.entries) }

// Contains returns true if the given group ID exists in the list.
func (l *ActiveGroupList) Contains(groupID int) bool {
	_, ok := l.entries[groupID]
	return ok
}

// Add adds a group ID to the list. Returns ErrActiveGroupLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the group ID already exists.
func (l *ActiveGroupList) Add(groupID int, maxActiveGroups int) error {
	if l.Contains(groupID) {
		return ErrDuplicateEntry
	}
	if len(l.entries) >= maxActiveGroups {
		return ErrActiveGroupLimitReached
	}
	l.entries[groupID] = struct{}{}
	return nil
}

// Remove removes a group ID from the list.
func (l *ActiveGroupList) Remove(groupID int) {
	delete(l.entries, groupID)
}
