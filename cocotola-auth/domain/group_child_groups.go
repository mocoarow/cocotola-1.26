package domain

import "errors"

// GroupChildGroups is an aggregate that manages the set of child group IDs belonging to a group.
type GroupChildGroups struct {
	groupID       int
	childGroupIDs map[int]struct{}
}

// NewGroupChildGroups creates a validated GroupChildGroups.
func NewGroupChildGroups(groupID int, childGroupIDs []int) (*GroupChildGroups, error) {
	if groupID <= 0 {
		return nil, errors.New("group child groups group id must be positive")
	}
	m := make(map[int]struct{}, len(childGroupIDs))
	for _, id := range childGroupIDs {
		m[id] = struct{}{}
	}
	return &GroupChildGroups{
		groupID:       groupID,
		childGroupIDs: m,
	}, nil
}

// GroupID returns the group ID.
func (g *GroupChildGroups) GroupID() int { return g.groupID }

// ChildGroupIDs returns a copy of the current child group IDs as a slice.
func (g *GroupChildGroups) ChildGroupIDs() []int {
	result := make([]int, 0, len(g.childGroupIDs))
	for id := range g.childGroupIDs {
		result = append(result, id)
	}
	return result
}

// Size returns the number of child groups.
func (g *GroupChildGroups) Size() int { return len(g.childGroupIDs) }

// Contains returns true if the given group ID exists in the child groups.
func (g *GroupChildGroups) Contains(groupID int) bool {
	_, ok := g.childGroupIDs[groupID]
	return ok
}

// Add adds a child group ID. Returns ErrDuplicateEntry if the group ID already exists.
func (g *GroupChildGroups) Add(childGroupID int) error {
	if g.Contains(childGroupID) {
		return ErrDuplicateEntry
	}
	g.childGroupIDs[childGroupID] = struct{}{}
	return nil
}

// Remove removes a child group ID.
func (g *GroupChildGroups) Remove(childGroupID int) {
	delete(g.childGroupIDs, childGroupID)
}
