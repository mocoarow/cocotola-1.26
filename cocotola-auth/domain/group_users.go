package domain

import "errors"

// GroupUsers is an aggregate that manages the set of user IDs belonging to a group.
type GroupUsers struct {
	groupID int
	userIDs map[int]struct{}
}

// NewGroupUsers creates a validated GroupUsers.
func NewGroupUsers(groupID int, userIDs []int) (*GroupUsers, error) {
	if groupID <= 0 {
		return nil, errors.New("group users group id must be positive")
	}
	m := make(map[int]struct{}, len(userIDs))
	for _, id := range userIDs {
		m[id] = struct{}{}
	}
	return &GroupUsers{
		groupID: groupID,
		userIDs: m,
	}, nil
}

// GroupID returns the group ID.
func (g *GroupUsers) GroupID() int { return g.groupID }

// UserIDs returns a copy of the current user IDs as a slice.
func (g *GroupUsers) UserIDs() []int {
	result := make([]int, 0, len(g.userIDs))
	for id := range g.userIDs {
		result = append(result, id)
	}
	return result
}

// Size returns the number of users in the group.
func (g *GroupUsers) Size() int { return len(g.userIDs) }

// Contains returns true if the given user ID exists in the group.
func (g *GroupUsers) Contains(userID int) bool {
	_, ok := g.userIDs[userID]
	return ok
}

// Add adds a user ID to the group. Returns ErrDuplicateEntry if the user ID already exists.
func (g *GroupUsers) Add(userID int) error {
	if g.Contains(userID) {
		return ErrDuplicateEntry
	}
	g.userIDs[userID] = struct{}{}
	return nil
}

// Remove removes a user ID from the group.
func (g *GroupUsers) Remove(userID int) {
	delete(g.userIDs, userID)
}
