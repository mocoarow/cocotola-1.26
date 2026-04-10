package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

var fixtureOrgID = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")

func newEdge(parent, child int) group.HierarchyEdge {
	return group.ReconstructHierarchyEdge(parent, child)
}

func Test_NewHierarchyEdge_shouldReturnEdge_whenValid(t *testing.T) {
	t.Parallel()

	// given / when
	edge, err := group.NewHierarchyEdge(1, 2)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, edge.ParentGroupID())
	assert.Equal(t, 2, edge.ChildGroupID())
}

func Test_NewHierarchyEdge_shouldReturnError_whenParentIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewHierarchyEdge(0, 2)

	// then
	require.Error(t, err)
}

func Test_NewHierarchyEdge_shouldReturnError_whenChildIDIsNegative(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewHierarchyEdge(1, -1)

	// then
	require.Error(t, err)
}

func Test_NewGroupHierarchy_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewHierarchy(domain.OrganizationID{}, nil)

	// then
	require.Error(t, err)
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_whenNoCycle(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, nil)

	// when
	err := h.AddEdge(1, 2)

	// then
	require.NoError(t, err)
	assert.Len(t, h.Edges(), 1)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenSelfLoop(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, nil)

	// when
	err := h.AddEdge(1, 1)

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenDirectCycle(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
	})

	// when
	err := h.AddEdge(2, 1)

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenIndirectCycle(t *testing.T) {
	t.Parallel()

	// given
	// A -> B -> C, then try C -> A
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
		newEdge(2, 3),
	})

	// when
	err := h.AddEdge(3, 1)

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
	})

	// when
	err := h.AddEdge(1, 2)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_whenMultipleBranches(t *testing.T) {
	t.Parallel()

	// given
	// A -> B, A -> C
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
	})

	// when
	err := h.AddEdge(1, 3)

	// then
	require.NoError(t, err)
	assert.Len(t, h.Edges(), 2)
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_whenDiamondShape(t *testing.T) {
	t.Parallel()

	// given
	// A -> B, A -> C, B -> D
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
		newEdge(1, 3),
		newEdge(2, 4),
	})

	// when - C -> D (diamond, not a cycle)
	err := h.AddEdge(3, 4)

	// then
	require.NoError(t, err)
}

func Test_GroupHierarchy_RemoveEdge_shouldRemoveEdge(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
		newEdge(2, 3),
	})

	// when
	h.RemoveEdge(1, 2)

	// then
	assert.Len(t, h.Edges(), 1)
	assert.Equal(t, 2, h.Edges()[0].ParentGroupID())
	assert.Equal(t, 3, h.Edges()[0].ChildGroupID())
}

func Test_GroupHierarchy_RemoveEdge_shouldDoNothing_whenEdgeNotFound(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
	})

	// when
	h.RemoveEdge(99, 100)

	// then
	assert.Len(t, h.Edges(), 1)
}

func Test_GroupHierarchy_RemoveGroup_shouldRemoveAllEdgesForGroup(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
		newEdge(2, 3),
		newEdge(4, 5),
	})

	// when
	h.RemoveGroup(2)

	// then
	assert.Len(t, h.Edges(), 1)
	assert.Equal(t, 4, h.Edges()[0].ParentGroupID())
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_afterCycleEdgeRemoved(t *testing.T) {
	t.Parallel()

	// given
	// A -> B -> C
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
		newEdge(2, 3),
	})

	// when - remove A -> B, then C -> A should succeed
	h.RemoveEdge(1, 2)
	err := h.AddEdge(3, 1)

	// then
	require.NoError(t, err)
}

func Test_GroupHierarchy_Edges_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(1, 2),
	})

	// when
	edges := h.Edges()
	edges[0] = newEdge(99, 100)

	// then
	assert.Equal(t, 1, h.Edges()[0].ParentGroupID())
}
