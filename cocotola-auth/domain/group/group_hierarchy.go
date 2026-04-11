package group

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// HierarchyEdge represents a parent-child relationship between groups.
type HierarchyEdge struct {
	parentGroupID domain.GroupID
	childGroupID  domain.GroupID
}

// NewHierarchyEdge creates a validated HierarchyEdge.
func NewHierarchyEdge(parentGroupID domain.GroupID, childGroupID domain.GroupID) (HierarchyEdge, error) {
	if parentGroupID.IsZero() {
		return HierarchyEdge{}, fmt.Errorf("hierarchy edge parent group id must not be zero: %w", domain.ErrInvalidArgument)
	}
	if childGroupID.IsZero() {
		return HierarchyEdge{}, fmt.Errorf("hierarchy edge child group id must not be zero: %w", domain.ErrInvalidArgument)
	}
	return HierarchyEdge{
		parentGroupID: parentGroupID,
		childGroupID:  childGroupID,
	}, nil
}

// ReconstructHierarchyEdge reconstitutes a HierarchyEdge from persistence.
func ReconstructHierarchyEdge(parentGroupID domain.GroupID, childGroupID domain.GroupID) HierarchyEdge {
	return HierarchyEdge{
		parentGroupID: parentGroupID,
		childGroupID:  childGroupID,
	}
}

// ParentGroupID returns the parent group ID.
func (e HierarchyEdge) ParentGroupID() domain.GroupID { return e.parentGroupID }

// ChildGroupID returns the child group ID.
func (e HierarchyEdge) ChildGroupID() domain.GroupID { return e.childGroupID }

// Hierarchy is an aggregate that manages all parent-child group relationships
// within an organization. It enforces the acyclic invariant by checking for cycles
// on every edge addition.
type Hierarchy struct {
	organizationID domain.OrganizationID
	edges          []HierarchyEdge
}

// NewHierarchy creates a validated GroupHierarchy.
func NewHierarchy(organizationID domain.OrganizationID, edges []HierarchyEdge) (*Hierarchy, error) {
	if organizationID.IsZero() {
		return nil, fmt.Errorf("group hierarchy organization id must not be zero: %w", domain.ErrInvalidArgument)
	}
	copied := make([]HierarchyEdge, len(edges))
	copy(copied, edges)
	return &Hierarchy{
		organizationID: organizationID,
		edges:          copied,
	}, nil
}

// OrganizationID returns the organization ID.
func (h *Hierarchy) OrganizationID() domain.OrganizationID { return h.organizationID }

// Edges returns a defensive copy of the current edges.
func (h *Hierarchy) Edges() []HierarchyEdge {
	copied := make([]HierarchyEdge, len(h.edges))
	copy(copied, h.edges)
	return copied
}

// AddEdge adds a parent-child edge. Returns ErrCyclicGroupHierarchy if the edge
// would create a cycle, or ErrDuplicateEntry if the edge already exists.
func (h *Hierarchy) AddEdge(parentGroupID domain.GroupID, childGroupID domain.GroupID) error {
	if parentGroupID.Equal(childGroupID) {
		return domain.ErrCyclicGroupHierarchy
	}
	for _, e := range h.edges {
		if e.parentGroupID.Equal(parentGroupID) && e.childGroupID.Equal(childGroupID) {
			return domain.ErrDuplicateEntry
		}
	}
	if h.hasPath(childGroupID, parentGroupID) {
		return domain.ErrCyclicGroupHierarchy
	}
	h.edges = append(h.edges, HierarchyEdge{
		parentGroupID: parentGroupID,
		childGroupID:  childGroupID,
	})
	return nil
}

// RemoveEdge removes a parent-child edge.
func (h *Hierarchy) RemoveEdge(parentGroupID domain.GroupID, childGroupID domain.GroupID) {
	filtered := make([]HierarchyEdge, 0, len(h.edges))
	for _, e := range h.edges {
		if !e.parentGroupID.Equal(parentGroupID) || !e.childGroupID.Equal(childGroupID) {
			filtered = append(filtered, e)
		}
	}
	h.edges = filtered
}

// RemoveGroup removes all edges involving the given group ID.
func (h *Hierarchy) RemoveGroup(groupID domain.GroupID) {
	filtered := make([]HierarchyEdge, 0, len(h.edges))
	for _, e := range h.edges {
		if !e.parentGroupID.Equal(groupID) && !e.childGroupID.Equal(groupID) {
			filtered = append(filtered, e)
		}
	}
	h.edges = filtered
}

// hasPath returns true if there is a directed path from `from` to `to`
// using BFS on the current edges.
func (h *Hierarchy) hasPath(from domain.GroupID, to domain.GroupID) bool {
	adj := make(map[domain.GroupID][]domain.GroupID)
	for _, e := range h.edges {
		adj[e.parentGroupID] = append(adj[e.parentGroupID], e.childGroupID)
	}

	visited := make(map[domain.GroupID]bool)
	queue := []domain.GroupID{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, child := range adj[current] {
			if child.Equal(to) {
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
