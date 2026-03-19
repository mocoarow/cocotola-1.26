package domain

import "errors"

// GroupUsers is an aggregate that manages the set of user IDs belonging to a group.
type GroupUsers struct{ intIDSet }

// NewGroupUsers creates a validated GroupUsers.
func NewGroupUsers(groupID int, userIDs []int) (*GroupUsers, error) {
	if groupID <= 0 {
		return nil, errors.New("group users group id must be positive")
	}
	return &GroupUsers{newIntIDSet(groupID, userIDs)}, nil
}

// GroupID returns the group ID.
func (g *GroupUsers) GroupID() int { return g.getOwnerID() }

// UserIDs returns a copy of the current user IDs as a slice.
func (g *GroupUsers) UserIDs() []int { return g.ids() }

// Add adds a user ID to the group. Returns ErrDuplicateEntry if the user ID already exists.
func (g *GroupUsers) Add(userID int) error { return g.addUnique(userID) }
