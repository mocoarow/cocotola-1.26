//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

func findGroupHierarchy(ctx context.Context, t *testing.T, repo *gateway.GroupHierarchyRepository, orgID domain.OrganizationID) *domaingroup.Hierarchy {
	t.Helper()
	hierarchy, err := repo.FindByOrganizationID(ctx, orgID)
	require.NoError(t, err)
	return hierarchy
}

func saveGroupEdges(ctx context.Context, t *testing.T, repo *gateway.GroupHierarchyRepository, orgID domain.OrganizationID, edges ...domaingroup.HierarchyEdge) {
	t.Helper()
	hierarchy := findGroupHierarchy(ctx, t, repo, orgID)
	for _, e := range edges {
		require.NoError(t, hierarchy.AddEdge(e.ParentGroupID(), e.ChildGroupID()))
	}
	require.NoError(t, repo.Save(ctx, hierarchy))
}

func Test_GroupHierarchyRepository_Save_shouldPersistEdges_whenFoundAgain(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-save-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-save-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	edges := []domaingroup.HierarchyEdge{
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
	}

	// when
	saveGroupEdges(ctx, t, repo, orgID, edges...)

	// then
	assert.ElementsMatch(t, edges, findGroupHierarchy(ctx, t, repo, orgID).Edges())
}

func Test_GroupHierarchyRepository_FindByOrganizationID_shouldReturnEmptyHierarchy_whenNoEdges(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-empty-org")
	repo := gateway.NewGroupHierarchyRepository(tx)

	// when
	found := findGroupHierarchy(ctx, t, repo, orgID)

	// then
	assert.Empty(t, found.Edges())
}

func Test_GroupHierarchyRepository_FindByOrganizationID_shouldReturnSavedVersion_whenSavedTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-version-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-version-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	saveGroupEdges(ctx, t, repo, orgID, domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]))
	saveGroupEdges(ctx, t, repo, orgID, domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]))

	// when
	found := findGroupHierarchy(ctx, t, repo, orgID)

	// then
	assert.Equal(t, 2, found.Version())
}

func Test_GroupHierarchyRepository_Save_shouldReplaceEdges_whenEdgeIsRemoved(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-replace-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-replace-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	removed := domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1])
	kept := domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2])
	saveGroupEdges(ctx, t, repo, orgID, removed, kept)
	hierarchy := findGroupHierarchy(ctx, t, repo, orgID)
	hierarchy.RemoveEdge(removed.ParentGroupID(), removed.ChildGroupID())

	// when
	err := repo.Save(ctx, hierarchy)

	// then
	require.NoError(t, err)
	assert.Equal(t, []domaingroup.HierarchyEdge{kept}, findGroupHierarchy(ctx, t, repo, orgID).Edges())
}

func Test_GroupHierarchyRepository_Save_shouldReturnErrConcurrentModification_whenRootIsCreatedConcurrently(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-first-save-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-first-save-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	first := findGroupHierarchy(ctx, t, repo, orgID)
	second := findGroupHierarchy(ctx, t, repo, orgID)
	require.NoError(t, first.AddEdge(groupIDs[0], groupIDs[1]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.AddEdge(groupIDs[0], groupIDs[2]))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_GroupHierarchyRepository_Save_shouldReturnErrConcurrentModification_whenVersionIsStale(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-stale-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-stale-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	saveGroupEdges(ctx, t, repo, orgID, domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]))
	first := findGroupHierarchy(ctx, t, repo, orgID)
	second := findGroupHierarchy(ctx, t, repo, orgID)
	require.NoError(t, first.AddEdge(groupIDs[1], groupIDs[2]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.AddEdge(groupIDs[2], groupIDs[0]))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_GroupHierarchyRepository_Save_shouldKeepEdgesOfBoth_whenLoserRetriesAfterConflict(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-both-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-both-org", 4)
	repo := gateway.NewGroupHierarchyRepository(tx)
	saveGroupEdges(ctx, t, repo, orgID, domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]))
	first := findGroupHierarchy(ctx, t, repo, orgID)
	second := findGroupHierarchy(ctx, t, repo, orgID)
	require.NoError(t, first.AddEdge(groupIDs[0], groupIDs[2]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.AddEdge(groupIDs[0], groupIDs[3]))
	require.ErrorIs(t, repo.Save(ctx, second), libversioned.ErrConcurrentModification)
	reread := findGroupHierarchy(ctx, t, repo, orgID)
	require.NoError(t, reread.AddEdge(groupIDs[0], groupIDs[3]))

	// when
	err := repo.Save(ctx, reread)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, []domaingroup.HierarchyEdge{
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[3]),
	}, findGroupHierarchy(ctx, t, repo, orgID).Edges())
}

func Test_GroupHierarchyRepository_FindByOrganizationID_shouldReturnGivenOrganizationID_whenEdgesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-orgid-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-orgid-org", 2)
	repo := gateway.NewGroupHierarchyRepository(tx)
	saveGroupEdges(ctx, t, repo, orgID, domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]))

	// when
	found := findGroupHierarchy(ctx, t, repo, orgID)

	// then
	assert.True(t, orgID.Equal(found.OrganizationID()))
}

func Test_GroupHierarchyRepository_FindByOrganizationID_shouldReturnHierarchyRejectingCycle_whenLoserRereadsAfterConflict(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "hierarchy-cycle-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-cycle-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	saveGroupEdges(ctx, t, repo, orgID, domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]))
	first := findGroupHierarchy(ctx, t, repo, orgID)
	second := findGroupHierarchy(ctx, t, repo, orgID)
	require.NoError(t, first.AddEdge(groupIDs[1], groupIDs[2]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.AddEdge(groupIDs[2], groupIDs[0]))
	require.ErrorIs(t, repo.Save(ctx, second), libversioned.ErrConcurrentModification)

	// when
	err := findGroupHierarchy(ctx, t, repo, orgID).AddEdge(groupIDs[2], groupIDs[0])

	// then
	require.ErrorIs(t, err, domain.ErrCyclicGroupHierarchy)
}
