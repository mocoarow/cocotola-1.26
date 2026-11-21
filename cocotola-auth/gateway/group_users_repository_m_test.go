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

func findGroupUsers(ctx context.Context, t *testing.T, repo *gateway.GroupUsersRepository, groupID domain.GroupID) *domaingroup.Users {
	t.Helper()
	gu, err := repo.FindByGroupID(ctx, groupID)
	require.NoError(t, err)
	return gu
}

func saveGroupUsers(ctx context.Context, t *testing.T, repo *gateway.GroupUsersRepository, groupID domain.GroupID, userIDs ...domain.AppUserID) {
	t.Helper()
	gu := findGroupUsers(ctx, t, repo, groupID)
	for _, id := range userIDs {
		require.NoError(t, gu.Add(id))
	}
	require.NoError(t, repo.Save(ctx, gu))
}

func Test_GroupUsersRepository_Save_shouldPersistUsers_whenFoundAgain(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-save-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-save-org", 3)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-save-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)

	// when
	saveGroupUsers(ctx, t, repo, groupIDs[0], userIDs...)

	// then
	assert.ElementsMatch(t, userIDs, findGroupUsers(ctx, t, repo, groupIDs[0]).UserIDs())
}

func Test_GroupUsersRepository_FindByGroupID_shouldReturnEmptyGroupUsers_whenNoEntries(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-empty-org")
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-empty-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)

	// when
	found := findGroupUsers(ctx, t, repo, groupIDs[0])

	// then
	assert.Equal(t, 0, found.Size())
}

func Test_GroupUsersRepository_FindByGroupID_shouldReturnSavedVersion_whenSavedTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-version-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-version-org", 2)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-version-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	saveGroupUsers(ctx, t, repo, groupIDs[0], userIDs[0])
	saveGroupUsers(ctx, t, repo, groupIDs[0], userIDs[1])

	// when
	found := findGroupUsers(ctx, t, repo, groupIDs[0])

	// then
	assert.Equal(t, 2, found.Version())
}

func Test_GroupUsersRepository_Save_shouldReplaceUsers_whenUserIsRemoved(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-replace-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-replace-org", 2)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-replace-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	saveGroupUsers(ctx, t, repo, groupIDs[0], userIDs...)
	gu := findGroupUsers(ctx, t, repo, groupIDs[0])
	gu.Remove(userIDs[0])

	// when
	err := repo.Save(ctx, gu)

	// then
	require.NoError(t, err)
	assert.Equal(t, userIDs[1:], findGroupUsers(ctx, t, repo, groupIDs[0]).UserIDs())
}

func Test_GroupUsersRepository_Save_shouldReturnErrConcurrentModification_whenRootIsCreatedConcurrently(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-first-save-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-first-save-org", 2)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-first-save-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	first := findGroupUsers(ctx, t, repo, groupIDs[0])
	second := findGroupUsers(ctx, t, repo, groupIDs[0])
	require.NoError(t, first.Add(userIDs[0]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(userIDs[1]))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_GroupUsersRepository_Save_shouldReturnErrConcurrentModification_whenVersionIsStale(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-stale-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-stale-org", 3)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-stale-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	saveGroupUsers(ctx, t, repo, groupIDs[0], userIDs[0])
	first := findGroupUsers(ctx, t, repo, groupIDs[0])
	second := findGroupUsers(ctx, t, repo, groupIDs[0])
	require.NoError(t, first.Add(userIDs[1]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(userIDs[2]))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_GroupUsersRepository_Save_shouldKeepUsersOfBoth_whenLoserRetriesAfterConflict(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-both-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-both-org", 3)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-both-org", 1)
	repo := gateway.NewGroupUsersRepository(tx)
	saveGroupUsers(ctx, t, repo, groupIDs[0], userIDs[0])
	first := findGroupUsers(ctx, t, repo, groupIDs[0])
	second := findGroupUsers(ctx, t, repo, groupIDs[0])
	require.NoError(t, first.Add(userIDs[1]))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(userIDs[2]))
	require.ErrorIs(t, repo.Save(ctx, second), libversioned.ErrConcurrentModification)
	reread := findGroupUsers(ctx, t, repo, groupIDs[0])
	require.NoError(t, reread.Add(userIDs[2]))

	// when
	err := repo.Save(ctx, reread)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, userIDs, findGroupUsers(ctx, t, repo, groupIDs[0]).UserIDs())
}

func Test_GroupUsersRepository_FindByGroupID_shouldReturnGivenGroupID_whenEntriesExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "groupusers-groupid-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "groupusers-groupid-org", 1)
	groupIDs := setupGroups(ctx, t, tx, orgID, "groupusers-groupid-org", 2)
	repo := gateway.NewGroupUsersRepository(tx)
	saveGroupUsers(ctx, t, repo, groupIDs[1], userIDs...)

	// when
	found := findGroupUsers(ctx, t, repo, groupIDs[1])

	// then
	assert.Equal(t, groupIDs[1], found.GroupID())
}
