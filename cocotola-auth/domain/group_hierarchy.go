package domain

import "errors"

// HierarchyEdge represents a parent-child relationship between groups.
type HierarchyEdge struct {
	parentGroupID int
	childGroupID  int
}

// NewHierarchyEdge creates a validated HierarchyEdge.
func NewHierarchyEdge(parentGroupID int, childGroupID int) (HierarchyEdge, error) {
	if parentGroupID <= 0 {
		return HierarchyEdge{}, errors.New("hierarchy edge parent group id must be positive")
	}
	if childGroupID <= 0 {
		return HierarchyEdge{}, errors.New("hierarchy edge child group id must be positive")
	}
	return HierarchyEdge{
		parentGroupID: parentGroupID,
		childGroupID:  childGroupID,
	}, nil
}

// ReconstructHierarchyEdge reconstitutes a HierarchyEdge from persistence.
func ReconstructHierarchyEdge(parentGroupID int, childGroupID int) HierarchyEdge {
	return HierarchyEdge{
		parentGroupID: parentGroupID,
		childGroupID:  childGroupID,
	}
}

// ParentGroupID returns the parent group ID.
func (e HierarchyEdge) ParentGroupID() int { return e.parentGroupID }

// ChildGroupID returns the child group ID.
func (e HierarchyEdge) ChildGroupID() int { return e.childGroupID }

// GroupHierarchy is an aggregate that manages all parent-child group relationships
// within an organization. It enforces the acyclic invariant by checking for cycles
// on every edge addition.
type GroupHierarchy struct {
	organizationID int
	edges          []HierarchyEdge
}

// NewGroupHierarchy creates a validated GroupHierarchy.
func NewGroupHierarchy(organizationID int, edges []HierarchyEdge) (*GroupHierarchy, error) {
	if organizationID <= 0 {
		return nil, errors.New("group hierarchy organization id must be positive")
	}
	copied := make([]HierarchyEdge, len(edges))
	copy(copied, edges)
	return &GroupHierarchy{
		organizationID: organizationID,
		edges:          copied,
	}, nil
}

// OrganizationID returns the organization ID.
func (h *GroupHierarchy) OrganizationID() int { return h.organizationID }

// Edges returns a defensive copy of the current edges.
func (h *GroupHierarchy) Edges() []HierarchyEdge {
	copied := make([]HierarchyEdge, len(h.edges))
	copy(copied, h.edges)
	return copied
}

// AddEdge adds a parent-child edge. Returns ErrCyclicGroupHierarchy if the edge
// would create a cycle, or ErrDuplicateEntry if the edge already exists.
func (h *GroupHierarchy) AddEdge(parentGroupID int, childGroupID int) error {
	if parentGroupID == childGroupID {
		return ErrCyclicGroupHierarchy
	}
	for _, e := range h.edges {
		if e.parentGroupID == parentGroupID && e.childGroupID == childGroupID {
			return ErrDuplicateEntry
		}
	}
	if h.hasPath(childGroupID, parentGroupID) {
		return ErrCyclicGroupHierarchy
	}
	h.edges = append(h.edges, HierarchyEdge{
		parentGroupID: parentGroupID,
		childGroupID:  childGroupID,
	})
	return nil
}

// RemoveEdge removes a parent-child edge.
func (h *GroupHierarchy) RemoveEdge(parentGroupID int, childGroupID int) {
	filtered := make([]HierarchyEdge, 0, len(h.edges))
	for _, e := range h.edges {
		if e.parentGroupID != parentGroupID || e.childGroupID != childGroupID {
			filtered = append(filtered, e)
		}
	}
	h.edges = filtered
}

// RemoveGroup removes all edges involving the given group ID.
func (h *GroupHierarchy) RemoveGroup(groupID int) {
	filtered := make([]HierarchyEdge, 0, len(h.edges))
	for _, e := range h.edges {
		if e.parentGroupID != groupID && e.childGroupID != groupID {
			filtered = append(filtered, e)
		}
	}
	h.edges = filtered
}

// hasPath returns true if there is a directed path from `from` to `to`
// using BFS on the current edges.
func (h *GroupHierarchy) hasPath(from int, to int) bool {
	adj := make(map[int][]int)
	for _, e := range h.edges {
		adj[e.parentGroupID] = append(adj[e.parentGroupID], e.childGroupID)
	}

	visited := make(map[int]bool)
	queue := []int{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, child := range adj[current] {
			if child == to {
				return true
			}
			if !visited[child] {
				visited[child] = true
				queue = append(queue, child)
			}
		}
	}
	return false
}
