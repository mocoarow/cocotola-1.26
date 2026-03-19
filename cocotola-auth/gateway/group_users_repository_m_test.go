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

func Test_GroupUsersRepository_Save_shouldInsertEntries_whenListIsNotEmpty(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "groupusers-save-org")
	userIDs := setupUsers(t, tx, ctx, orgID, "groupusers-save-org", 2)
	groupIDs := setupGroups(t, tx, ctx, orgID, "groupusers-save-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	gu, err := domain.NewGroupUsers(groupIDs[0], userIDs)
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, gu)

	// then
	require.NoError(t, err)
}

func Test_GroupUsersRepository_FindByGroupID_shouldReturnGroupUsers_whenEntriesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "groupusers-find-org")
	userIDs := setupUsers(t, tx, ctx, orgID, "groupusers-find-org", 3)
	groupIDs := setupGroups(t, tx, ctx, orgID, "groupusers-find-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	gu, err := domain.NewGroupUsers(groupIDs[0], userIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, gu))

	// when
	found, err := repo.FindByGroupID(ctx, groupIDs[0])

	// then
	require.NoError(t, err)
	assert.Equal(t, groupIDs[0], found.GroupID())
	assert.Equal(t, 3, found.Size())
	for _, id := range userIDs {
		assert.True(t, found.Contains(id))
	}
}

func Test_GroupUsersRepository_FindByGroupID_shouldReturnEmptyGroupUsers_whenNoEntries(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "groupusers-empty-org")
	groupIDs := setupGroups(t, tx, ctx, orgID, "groupusers-empty-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)

	// when
	found, err := repo.FindByGroupID(ctx, groupIDs[0])

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, found.Size())
}

func Test_GroupUsersRepository_Save_shouldReplaceEntries_whenCalledTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(t, tx, ctx, "groupusers-replace-org")
	userIDs := setupUsers(t, tx, ctx, orgID, "groupusers-replace-org", 3)
	groupIDs := setupGroups(t, tx, ctx, orgID, "groupusers-replace-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)

	gu1, err := domain.NewGroupUsers(groupIDs[0], userIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, gu1))

	// when: save with only the first user
	gu2, err := domain.NewGroupUsers(groupIDs[0], userIDs[:1])
	require.NoError(t, err)
	err = repo.Save(ctx, gu2)

	// then
	require.NoError(t, err)
	found, err := repo.FindByGroupID(ctx, groupIDs[0])
	require.NoError(t, err)
	assert.Equal(t, 1, found.Size())
	assert.True(t, found.Contains(userIDs[0]))
}
