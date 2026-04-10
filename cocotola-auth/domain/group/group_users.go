package group

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// Users is an aggregate that manages the set of user IDs belonging to a group.
type Users struct {
	memberSetBase[domain.AppUserID]
}

// NewUsers creates a validated Users aggregate.
func NewUsers(groupID domain.GroupID, userIDs []domain.AppUserID) (*Users, error) {
	if groupID.IsZero() {
		return nil, errors.New("group users group id must not be zero")
	}
	return &Users{memberSetBase[domain.AppUserID]{idset.New[domain.GroupID, domain.AppUserID](groupID, userIDs)}}, nil
}

// UserIDs returns a copy of the current user IDs as a slice.
func (g *Users) UserIDs() []domain.AppUserID { return g.set.IDs() }

// Add adds a user ID to the group. Returns ErrDuplicateEntry if the user ID already exists.
func (g *Users) Add(userID domain.AppUserID) error {
	if err := g.set.AddUnique(userID, domain.ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add user to group: %w", err)
	}
	return nil
}
