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

func Test_ActiveGroupListRepository_Save_shouldInsertEntries_whenListIsNotEmpty(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activegroup-save-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "activegroup-save-org", 2)
	repo := gateway.NewActiveGroupListRepository(tx)
	list, err := domain.NewActiveGroupList(orgID, groupIDs)
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, list)

	// then
	require.NoError(t, err)
}

func Test_ActiveGroupListRepository_FindByOrganizationID_shouldReturnList_whenEntriesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activegroup-find-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "activegroup-find-org", 3)
	repo := gateway.NewActiveGroupListRepository(tx)
	list, err := domain.NewActiveGroupList(orgID, groupIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, list))

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, orgID, found.OrganizationID())
	assert.Equal(t, 3, found.Size())
	for _, id := range groupIDs {
		assert.True(t, found.Contains(id))
	}
}

func Test_ActiveGroupListRepository_FindByOrganizationID_shouldReturnEmptyList_whenNoEntries(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activegroup-empty-org")
	repo := gateway.NewActiveGroupListRepository(tx)

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, found.Size())
}

func Test_ActiveGroupListRepository_Save_shouldReplaceEntries_whenCalledTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activegroup-replace-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "activegroup-replace-org", 3)
	repo := gateway.NewActiveGroupListRepository(tx)

	list1, err := domain.NewActiveGroupList(orgID, groupIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, list1))

	// when: save with only the first group
	list2, err := domain.NewActiveGroupList(orgID, groupIDs[:1])
	require.NoError(t, err)
	err = repo.Save(ctx, list2)

	// then
	require.NoError(t, err)
	found, err := repo.FindByOrganizationID(ctx, orgID)
	require.NoError(t, err)
	assert.Equal(t, 1, found.Size())
	assert.True(t, found.Contains(groupIDs[0]))
}
