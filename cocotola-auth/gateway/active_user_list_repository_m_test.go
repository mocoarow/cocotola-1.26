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

func Test_ActiveUserListRepository_Save_shouldInsertEntries_whenListIsNotEmpty(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activeuser-save-org")
	userIDs := setupUsers(t, tx, ctx, orgID, "activeuser-save-org", 2)
	repo := gateway.NewActiveUserListRepository(tx)
	list, err := domain.NewActiveUserList(orgID, userIDs)
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, list)

	// then
	require.NoError(t, err)
}

func Test_ActiveUserListRepository_FindByOrganizationID_shouldReturnList_whenEntriesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activeuser-find-org")
	userIDs := setupUsers(t, tx, ctx, orgID, "activeuser-find-org", 3)
	repo := gateway.NewActiveUserListRepository(tx)
	list, err := domain.NewActiveUserList(orgID, userIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, list))

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, orgID, found.OrganizationID())
	assert.Equal(t, 3, found.Size())
	for _, id := range userIDs {
		assert.True(t, found.Contains(id))
	}
}

func Test_ActiveUserListRepository_FindByOrganizationID_shouldReturnEmptyList_whenNoEntries(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activeuser-empty-org")
	repo := gateway.NewActiveUserListRepository(tx)

	// when
	found, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, found.Size())
}

func Test_ActiveUserListRepository_Save_shouldReplaceEntries_whenCalledTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "activeuser-replace-org")
	userIDs := setupUsers(t, tx, ctx, orgID, "activeuser-replace-org", 3)
	repo := gateway.NewActiveUserListRepository(tx)

	list1, err := domain.NewActiveUserList(orgID, userIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, list1))

	// when: save with only the first user
	list2, err := domain.NewActiveUserList(orgID, userIDs[:1])
	require.NoError(t, err)
	err = repo.Save(ctx, list2)

	// then
	require.NoError(t, err)
	found, err := repo.FindByOrganizationID(ctx, orgID)
	require.NoError(t, err)
	assert.Equal(t, 1, found.Size())
	assert.True(t, found.Contains(userIDs[0]))
}
