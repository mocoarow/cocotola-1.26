package domain

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// ActiveUserList is an aggregate that manages the set of active user IDs for an organization,
// enforcing the "at most maxActiveUsers" invariant.
type ActiveUserList struct {
	activeListBase[AppUserID]
}

// NewActiveUserList creates a validated ActiveUserList.
func NewActiveUserList(organizationID OrganizationID, entries []AppUserID) (*ActiveUserList, error) {
	if organizationID.IsZero() {
		return nil, errors.New("active user list organization id must not be zero")
	}
	return &ActiveUserList{activeListBase[AppUserID]{idset.New[OrganizationID, AppUserID](organizationID, entries)}}, nil
}

// Add adds a user ID to the list. Returns ErrActiveUserLimitReached if the list is at capacity,
// or ErrDuplicateEntry if the user ID already exists.
func (l *ActiveUserList) Add(userID AppUserID, maxActiveUsers int) error {
	if err := l.set.AddWithLimit(userID, maxActiveUsers, ErrActiveUserLimitReached, ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add user to active list: %w", err)
	}
	return nil
}
