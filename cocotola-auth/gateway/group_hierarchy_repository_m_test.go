//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func Test_GroupHierarchyRepository_Save_shouldInsertEdges_whenEdgesExist(t *testing.T) {
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
	hierarchy, err := domaingroup.NewHierarchy(orgID, edges)
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
	orgID := setupOrganization(ctx, t, tx, "hierarchy-find-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-find-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)
	edges := []domaingroup.HierarchyEdge{
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
	}
	hierarchy, err := domaingroup.NewHierarchy(orgID, edges)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, hierarchy))

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.True(t, orgID.Equal(found.OrganizationID()))
	foundEdges := found.Edges()
	assert.Len(t, foundEdges, 2)
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
	orgID := setupOrganization(ctx, t, tx, "hierarchy-replace-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "hierarchy-replace-org", 3)
	repo := gateway.NewGroupHierarchyRepository(tx)

	edges1 := []domaingroup.HierarchyEdge{
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[1]),
		domaingroup.ReconstructHierarchyEdge(groupIDs[0], groupIDs[2]),
	}
	h1, err := domaingroup.NewHierarchy(orgID, edges1)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, h1))

	// when: save with only one edge
	edges2 := []domaingroup.HierarchyEdge{
		domaingroup.ReconstructHierarchyEdge(groupIDs[1], groupIDs[2]),
	}
	h2, err := domaingroup.NewHierarchy(orgID, edges2)
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
