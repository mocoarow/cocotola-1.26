package domain

import "errors"

// ActiveUserList is an aggregate that manages the set of active user IDs for an organization,
// enforcing the "at most maxActiveUsers" invariant.
type ActiveUserList struct{ intIDSet }

// NewActiveUserList creates a validated ActiveUserList.
func NewActiveUserList(organizationID int, entries []int) (*ActiveUserList, error) {
	if organizationID <= 0 {
		return nil, errors.New("active user list organization id must be positive")
	}
	return &ActiveUserList{newIntIDSet(organizationID, entries)}, nil
}

// OrganizationID returns the organization ID.
func (l *ActiveUserList) OrganizationID() int { return l.getOwnerID() }

// Entries returns a copy of the current user IDs as a slice.
func (l *ActiveUserList) Entries() []int { return l.ids() }

// Add adds a user ID to the list. Returns ErrActiveUserLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the user ID already exists.
func (l *ActiveUserList) Add(userID int, maxActiveUsers int) error {
	return l.addWithLimit(userID, maxActiveUsers, ErrActiveUserLimitReached)
}
