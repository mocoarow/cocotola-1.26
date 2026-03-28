package group

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// Users is an aggregate that manages the set of user IDs belonging to a group.
type Users struct{ memberSetBase }

// NewUsers creates a validated Users aggregate.
func NewUsers(groupID int, userIDs []int) (*Users, error) {
	if groupID <= 0 {
		return nil, errors.New("group users group id must be positive")
	}
	return &Users{memberSetBase{idset.New(groupID, userIDs)}}, nil
}

// UserIDs returns a copy of the current user IDs as a slice.
func (g *Users) UserIDs() []int { return g.set.IDs() }

// Add adds a user ID to the group. Returns ErrDuplicateEntry if the user ID already exists.
func (g *Users) Add(userID int) error {
	if err := g.set.AddUnique(userID, domain.ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add user to group: %w", err)
	}
	return nil
}
