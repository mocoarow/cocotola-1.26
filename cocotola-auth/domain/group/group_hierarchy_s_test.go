package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

var (
	fixtureOrgID = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")

	fixtureHierGroupID1   = domain.MustParseGroupID("00000000-0000-7000-8000-000000000001")
	fixtureHierGroupID2   = domain.MustParseGroupID("00000000-0000-7000-8000-000000000002")
	fixtureHierGroupID3   = domain.MustParseGroupID("00000000-0000-7000-8000-000000000003")
	fixtureHierGroupID4   = domain.MustParseGroupID("00000000-0000-7000-8000-000000000004")
	fixtureHierGroupID5   = domain.MustParseGroupID("00000000-0000-7000-8000-000000000005")
	fixtureHierGroupID99  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000099")
	fixtureHierGroupID100 = domain.MustParseGroupID("00000000-0000-7000-8000-000000000100")
)

func newEdge(parent, child domain.GroupID) group.HierarchyEdge {
	return group.ReconstructHierarchyEdge(parent, child)
}

func Test_NewHierarchyEdge_shouldReturnEdge_whenValid(t *testing.T) {
	t.Parallel()

	// given / when
	edge, err := group.NewHierarchyEdge(fixtureHierGroupID1, fixtureHierGroupID2)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureHierGroupID1, edge.ParentGroupID())
	assert.Equal(t, fixtureHierGroupID2, edge.ChildGroupID())
}

func Test_NewHierarchyEdge_shouldReturnError_whenParentIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewHierarchyEdge(domain.GroupID{}, fixtureHierGroupID2)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewHierarchyEdge_shouldReturnError_whenChildIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewHierarchyEdge(fixtureHierGroupID1, domain.GroupID{})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewGroupHierarchy_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewHierarchy(domain.OrganizationID{}, nil)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_whenNoCycle(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, nil)

	// when
	err := h.AddEdge(fixtureHierGroupID1, fixtureHierGroupID2)

	// then
	require.NoError(t, err)
	assert.Len(t, h.Edges(), 1)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenSelfLoop(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, nil)

	// when
	err := h.AddEdge(fixtureHierGroupID1, fixtureHierGroupID1)

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenDirectCycle(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
	})

	// when
	err := h.AddEdge(fixtureHierGroupID2, fixtureHierGroupID1)

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenIndirectCycle(t *testing.T) {
	t.Parallel()

	// given
	// A -> B -> C, then try C -> A
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
		newEdge(fixtureHierGroupID2, fixtureHierGroupID3),
	})

	// when
	err := h.AddEdge(fixtureHierGroupID3, fixtureHierGroupID1)

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}

func Test_GroupHierarchy_AddEdge_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
	})

	// when
	err := h.AddEdge(fixtureHierGroupID1, fixtureHierGroupID2)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_whenMultipleBranches(t *testing.T) {
	t.Parallel()

	// given
	// A -> B, A -> C
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
	})

	// when
	err := h.AddEdge(fixtureHierGroupID1, fixtureHierGroupID3)

	// then
	require.NoError(t, err)
	assert.Len(t, h.Edges(), 2)
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_whenDiamondShape(t *testing.T) {
	t.Parallel()

	// given
	// A -> B, A -> C, B -> D
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
		newEdge(fixtureHierGroupID1, fixtureHierGroupID3),
		newEdge(fixtureHierGroupID2, fixtureHierGroupID4),
	})

	// when - C -> D (diamond, not a cycle)
	err := h.AddEdge(fixtureHierGroupID3, fixtureHierGroupID4)

	// then
	require.NoError(t, err)
}

func Test_GroupHierarchy_RemoveEdge_shouldRemoveEdge(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
		newEdge(fixtureHierGroupID2, fixtureHierGroupID3),
	})

	// when
	h.RemoveEdge(fixtureHierGroupID1, fixtureHierGroupID2)

	// then
	assert.Len(t, h.Edges(), 1)
	assert.Equal(t, fixtureHierGroupID2, h.Edges()[0].ParentGroupID())
	assert.Equal(t, fixtureHierGroupID3, h.Edges()[0].ChildGroupID())
}

func Test_GroupHierarchy_RemoveEdge_shouldDoNothing_whenEdgeNotFound(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
	})

	// when
	h.RemoveEdge(fixtureHierGroupID99, fixtureHierGroupID100)

	// then
	assert.Len(t, h.Edges(), 1)
}

func Test_GroupHierarchy_RemoveGroup_shouldRemoveAllEdgesForGroup(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
		newEdge(fixtureHierGroupID2, fixtureHierGroupID3),
		newEdge(fixtureHierGroupID4, fixtureHierGroupID5),
	})

	// when
	h.RemoveGroup(fixtureHierGroupID2)

	// then
	assert.Len(t, h.Edges(), 1)
	assert.Equal(t, fixtureHierGroupID4, h.Edges()[0].ParentGroupID())
}

func Test_GroupHierarchy_AddEdge_shouldSucceed_afterCycleEdgeRemoved(t *testing.T) {
	t.Parallel()

	// given
	// A -> B -> C
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
		newEdge(fixtureHierGroupID2, fixtureHierGroupID3),
	})

	// when - remove A -> B, then C -> A should succeed
	h.RemoveEdge(fixtureHierGroupID1, fixtureHierGroupID2)
	err := h.AddEdge(fixtureHierGroupID3, fixtureHierGroupID1)

	// then
	require.NoError(t, err)
}

func Test_GroupHierarchy_Edges_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	h, _ := group.NewHierarchy(fixtureOrgID, []group.HierarchyEdge{
		newEdge(fixtureHierGroupID1, fixtureHierGroupID2),
	})

	// when
	edges := h.Edges()
	edges[0] = newEdge(fixtureHierGroupID99, fixtureHierGroupID100)

	// then
	assert.Equal(t, fixtureHierGroupID1, h.Edges()[0].ParentGroupID())
}
