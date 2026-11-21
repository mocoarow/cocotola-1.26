package group

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"
)

// Users is an aggregate that manages the set of user IDs belonging to a group.
type Users struct {
	memberSetBase[domain.AppUserID]
	version int
}

// NewUsers creates a validated Users aggregate.
func NewUsers(groupID domain.GroupID, userIDs []domain.AppUserID) (*Users, error) {
	if groupID.IsZero() {
		return nil, fmt.Errorf("group users group id must not be zero: %w", domain.ErrInvalidArgument)
	}
	return &Users{memberSetBase: memberSetBase[domain.AppUserID]{idset.New[domain.GroupID, domain.AppUserID](groupID, userIDs)}, version: 0}, nil
}

// Version returns the persisted version (0 = not yet saved).
func (g *Users) Version() int { return g.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
func (g *Users) SetVersion(version int) { g.version = version }

// UserIDs returns a copy of the current user IDs as a slice.
func (g *Users) UserIDs() []domain.AppUserID { return g.set.IDs() }

// Add adds a user ID to the group. Returns ErrDuplicateEntry if the user ID already exists.
func (g *Users) Add(userID domain.AppUserID) error {
	if err := g.set.AddUnique(userID, domain.ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add user to group: %w", err)
	}
	return nil
}
