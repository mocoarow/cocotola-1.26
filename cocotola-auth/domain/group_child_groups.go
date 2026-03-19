package domain

import "errors"

// GroupChildGroups is an aggregate that manages the set of child group IDs belonging to a group.
type GroupChildGroups struct{ intIDSet }

// NewGroupChildGroups creates a validated GroupChildGroups.
func NewGroupChildGroups(groupID int, childGroupIDs []int) (*GroupChildGroups, error) {
	if groupID <= 0 {
		return nil, errors.New("group child groups group id must be positive")
	}
	return &GroupChildGroups{newIntIDSet(groupID, childGroupIDs)}, nil
}

// GroupID returns the group ID.
func (g *GroupChildGroups) GroupID() int { return g.getOwnerID() }

// ChildGroupIDs returns a copy of the current child group IDs as a slice.
func (g *GroupChildGroups) ChildGroupIDs() []int { return g.ids() }

// Add adds a child group ID. Returns ErrDuplicateEntry if the group ID already exists.
func (g *GroupChildGroups) Add(childGroupID int) error { return g.addUnique(childGroupID) }
