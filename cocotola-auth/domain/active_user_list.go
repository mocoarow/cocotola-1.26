package domain

import "errors"

// ActiveUserList is an aggregate that manages the set of active user IDs for an organization,
// enforcing the "at most maxActiveUsers" invariant.
type ActiveUserList struct {
	organizationID int
	entries        map[int]struct{}
}

// NewActiveUserList creates a validated ActiveUserList.
func NewActiveUserList(organizationID int, entries []int) (*ActiveUserList, error) {
	if organizationID <= 0 {
		return nil, errors.New("active user list organization id must be positive")
	}
	m := make(map[int]struct{}, len(entries))
	for _, id := range entries {
		m[id] = struct{}{}
	}
	return &ActiveUserList{
		organizationID: organizationID,
		entries:        m,
	}, nil
}

// OrganizationID returns the organization ID.
func (l *ActiveUserList) OrganizationID() int { return l.organizationID }

// Entries returns a copy of the current user IDs as a slice.
func (l *ActiveUserList) Entries() []int {
	result := make([]int, 0, len(l.entries))
	for id := range l.entries {
		result = append(result, id)
	}
	return result
}

// Size returns the number of active users.
func (l *ActiveUserList) Size() int { return len(l.entries) }

// Contains returns true if the given user ID exists in the list.
func (l *ActiveUserList) Contains(userID int) bool {
	_, ok := l.entries[userID]
	return ok
}

// Add adds a user ID to the list. Returns ErrActiveUserLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the user ID already exists.
func (l *ActiveUserList) Add(userID int, maxActiveUsers int) error {
	if l.Contains(userID) {
		return ErrDuplicateEntry
	}
	if len(l.entries) >= maxActiveUsers {
		return ErrActiveUserLimitReached
	}
	l.entries[userID] = struct{}{}
	return nil
}

// Remove removes a user ID from the list.
func (l *ActiveUserList) Remove(userID int) {
	delete(l.entries, userID)
}
