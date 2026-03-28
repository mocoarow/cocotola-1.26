package domain

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// ActiveUserList is an aggregate that manages the set of active user IDs for an organization,
// enforcing the "at most maxActiveUsers" invariant.
type ActiveUserList struct{ activeListBase }

// NewActiveUserList creates a validated ActiveUserList.
func NewActiveUserList(organizationID int, entries []int) (*ActiveUserList, error) {
	if organizationID <= 0 {
		return nil, errors.New("active user list organization id must be positive")
	}
	return &ActiveUserList{activeListBase{idset.New(organizationID, entries)}}, nil
}

// Add adds a user ID to the list. Returns ErrActiveUserLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the user ID already exists.
func (l *ActiveUserList) Add(userID int, maxActiveUsers int) error {
	if err := l.set.AddWithLimit(userID, maxActiveUsers, ErrActiveUserLimitReached, ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add user to active list: %w", err)
	}
	return nil
}
