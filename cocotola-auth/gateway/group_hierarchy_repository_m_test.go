//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func Test_GroupHierarchyRepository_Save_shouldInsertEdges_whenEdgesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "hierarchy-save-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "hierarchy-save-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	edges := []domain.HierarchyEdge{
		domain.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domain.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
	}
	hierarchy, err := domain.NewGroupHierarchy(orgID, edges)
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, hierarchy)

	// then
	require.NoError(t, err)
}

func Test_GroupHierarchyRepository_FindByOrganizationID_shouldReturnHierarchy_whenEdgesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "hierarchy-find-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "hierarchy-find-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	edges := []domain.HierarchyEdge{
		domain.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domain.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
	}
	hierarchy, err := domain.NewGroupHierarchy(orgID, edges)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, hierarchy))

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, orgID, found.OrganizationID())
	foundEdges := found.Edges()
	assert.Len(t, foundEdges, 2)
}

func Test_GroupHierarchyRepository_FindByOrganizationID_shouldReturnEmptyHierarchy_whenNoEdges(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "hierarchy-empty-org")
	repo := gateway.NewGroupHierarchyRepository(tx)

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Empty(t, found.Edges())
}

func Test_GroupHierarchyRepository_Save_shouldReplaceEdges_whenCalledTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "hierarchy-replace-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "hierarchy-replace-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)

	edges1 := []domain.HierarchyEdge{
		domain.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domain.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
	}
	h1, err := domain.NewGroupHierarchy(orgID, edges1)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, h1))

	// when: save with only one edge
	edges2 := []domain.HierarchyEdge{
		domain.ReconstructHierarchyEdge(groupIDs[1], groupIDs[2]),
	}
	h2, err := domain.NewGroupHierarchy(orgID, edges2)
	require.NoError(t, err)
	err = repo.Save(ctx, h2)

	// then
	require.NoError(t, err)
	found, err := repo.FindByOrganizationID(ctx, orgID)
	require.NoError(t, err)
	foundEdges := found.Edges()
	assert.Len(t, foundEdges, 1)
	assert.Equal(t, groupIDs[1], foundEdges[0].ParentGroupID())
	assert.Equal(t, groupIDs[2], foundEdges[0].ChildGroupID())
}
