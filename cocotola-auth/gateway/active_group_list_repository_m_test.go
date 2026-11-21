//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

func findActiveGroupList(ctx context.Context, t *testing.T, repo *gateway.ActiveGroupListRepository, orgID domain.OrganizationID) *domain.ActiveGroupList {
	t.Helper()
	list, err := repo.FindByOrganizationID(ctx, orgID)
	require.NoError(t, err)
	return list
}

func saveActiveGroups(ctx context.Context, t *testing.T, repo *gateway.ActiveGroupListRepository, orgID domain.OrganizationID, groupIDs ...domain.GroupID) {
	t.Helper()
	list := findActiveGroupList(ctx, t, repo, orgID)
	for _, id := range groupIDs {
		require.NoError(t, list.Add(id, activeListLimit))
	}
	require.NoError(t, repo.Save(ctx, list))
}

func Test_ActiveGroupListRepository_Save_shouldPersistEntries_whenFoundAgain(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activegroup-save-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "activegroup-save-org", 2)
	repo := gateway.NewActiveGroupListRepository(tx)

	// when
	saveActiveGroups(ctx, t, repo, orgID, groupIDs...)

	// then
	assert.ElementsMatch(t, groupIDs, findActiveGroupList(ctx, t, repo, orgID).Entries())
}

func Test_ActiveGroupListRepository_FindByOrganizationID_shouldReturnEmptyList_whenNoEntries(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activegroup-empty-org")
	repo := gateway.NewActiveGroupListRepository(tx)

	// when
	found := findActiveGroupList(ctx, t, repo, orgID)

	// then
	assert.Equal(t, 0, found.Size())
}

func Test_ActiveGroupListRepository_FindByOrganizationID_shouldReturnSavedVersion_whenSavedTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activegroup-version-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "activegroup-version-org", 2)
	repo := gateway.NewActiveGroupListRepository(tx)
	saveActiveGroups(ctx, t, repo, orgID, groupIDs[0])
	saveActiveGroups(ctx, t, repo, orgID, groupIDs[1])

	// when
	found := findActiveGroupList(ctx, t, repo, orgID)

	// then
	assert.Equal(t, 2, found.Version())
}

func Test_ActiveGroupListRepository_Save_shouldReplaceEntries_whenEntryIsRemoved(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activegroup-replace-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "activegroup-replace-org", 2)
	repo := gateway.NewActiveGroupListRepository(tx)
	saveActiveGroups(ctx, t, repo, orgID, groupIDs...)
	list := findActiveGroupList(ctx, t, repo, orgID)
	list.Remove(groupIDs[0])

	// when
	err := repo.Save(ctx, list)

	// then
	require.NoError(t, err)
	assert.Equal(t, groupIDs[1:], findActiveGroupList(ctx, t, repo, orgID).Entries())
}

func Test_ActiveGroupListRepository_Save_shouldReturnErrConcurrentModification_whenRootIsCreatedConcurrently(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activegroup-first-save-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "activegroup-first-save-org", 2)
	repo := gateway.NewActiveGroupListRepository(tx)
	first := findActiveGroupList(ctx, t, repo, orgID)
	second := findActiveGroupList(ctx, t, repo, orgID)
	require.NoError(t, first.Add(groupIDs[0], activeListLimit))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(groupIDs[1], activeListLimit))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_ActiveGroupListRepository_Save_shouldReturnErrConcurrentModification_whenVersionIsStale(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activegroup-stale-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "activegroup-stale-org", 3)
	repo := gateway.NewActiveGroupListRepository(tx)
	saveActiveGroups(ctx, t, repo, orgID, groupIDs[0])
	first := findActiveGroupList(ctx, t, repo, orgID)
	second := findActiveGroupList(ctx, t, repo, orgID)
	require.NoError(t, first.Add(groupIDs[1], activeListLimit))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(groupIDs[2], activeListLimit))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}
