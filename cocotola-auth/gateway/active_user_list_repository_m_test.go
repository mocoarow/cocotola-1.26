//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

const activeListLimit = 5

func findActiveUserList(ctx context.Context, t *testing.T, repo *gateway.ActiveUserListRepository, orgID domain.OrganizationID) *domain.ActiveUserList {
	t.Helper()
	list, err := repo.FindByOrganizationID(ctx, orgID)
	require.NoError(t, err)
	return list
}

func saveActiveUsers(ctx context.Context, t *testing.T, repo *gateway.ActiveUserListRepository, orgID domain.OrganizationID, userIDs ...domain.AppUserID) {
	t.Helper()
	list := findActiveUserList(ctx, t, repo, orgID)
	for _, id := range userIDs {
		require.NoError(t, list.Add(id, activeListLimit))
	}
	require.NoError(t, repo.Save(ctx, list))
}

func givenLegacyActiveUsers(ctx context.Context, t *testing.T, tx *gorm.DB, name string) (domain.OrganizationID, []domain.AppUserID) {
	t.Helper()
	orgID := setupOrganization(ctx, t, tx, name)
	userIDs := setupUsers(ctx, t, tx, orgID, name, 2)
	for _, id := range userIDs {
		require.NoError(t, tx.Exec("INSERT INTO active_user (organization_id, user_id) VALUES (?, ?)", orgID.String(), id.String()).Error)
	}
	return orgID, userIDs
}

func Test_ActiveUserListRepository_Save_shouldPersistEntries_whenFoundAgain(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-save-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "activeuser-save-org", 2)
	repo := gateway.NewActiveUserListRepository(tx)

	// when
	saveActiveUsers(ctx, t, repo, orgID, userIDs...)

	// then
	assert.ElementsMatch(t, userIDs, findActiveUserList(ctx, t, repo, orgID).Entries())
}

func Test_ActiveUserListRepository_FindByOrganizationID_shouldReturnEmptyList_whenNoEntries(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-empty-org")
	repo := gateway.NewActiveUserListRepository(tx)

	// when
	found := findActiveUserList(ctx, t, repo, orgID)

	// then
	assert.Equal(t, 0, found.Size())
}

func Test_ActiveUserListRepository_FindByOrganizationID_shouldReturnSavedVersion_whenSavedTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-version-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "activeuser-version-org", 2)
	repo := gateway.NewActiveUserListRepository(tx)
	saveActiveUsers(ctx, t, repo, orgID, userIDs[0])
	saveActiveUsers(ctx, t, repo, orgID, userIDs[1])

	// when
	found := findActiveUserList(ctx, t, repo, orgID)

	// then
	assert.Equal(t, 2, found.Version())
}

func Test_ActiveUserListRepository_Save_shouldReplaceEntries_whenEntryIsRemoved(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-replace-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "activeuser-replace-org", 2)
	repo := gateway.NewActiveUserListRepository(tx)
	saveActiveUsers(ctx, t, repo, orgID, userIDs...)
	list := findActiveUserList(ctx, t, repo, orgID)
	list.Remove(userIDs[0])

	// when
	err := repo.Save(ctx, list)

	// then
	require.NoError(t, err)
	assert.Equal(t, userIDs[1:], findActiveUserList(ctx, t, repo, orgID).Entries())
}

func Test_ActiveUserListRepository_Save_shouldReturnErrConcurrentModification_whenVersionIsStale(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-stale-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "activeuser-stale-org", 3)
	repo := gateway.NewActiveUserListRepository(tx)
	saveActiveUsers(ctx, t, repo, orgID, userIDs[0])
	first := findActiveUserList(ctx, t, repo, orgID)
	second := findActiveUserList(ctx, t, repo, orgID)
	require.NoError(t, first.Add(userIDs[1], activeListLimit))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(userIDs[2], activeListLimit))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_ActiveUserListRepository_Save_shouldReturnErrConcurrentModification_whenRootIsCreatedConcurrently(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-first-save-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "activeuser-first-save-org", 2)
	repo := gateway.NewActiveUserListRepository(tx)
	first := findActiveUserList(ctx, t, repo, orgID)
	second := findActiveUserList(ctx, t, repo, orgID)
	require.NoError(t, first.Add(userIDs[0], activeListLimit))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(userIDs[1], activeListLimit))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_ActiveUserListRepository_FindByOrganizationID_shouldReturnLegacyEntries_whenRootRowIsMissing(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID, userIDs := givenLegacyActiveUsers(ctx, t, tx, "activeuser-legacy-entries-org")
	repo := gateway.NewActiveUserListRepository(tx)

	// when
	found := findActiveUserList(ctx, t, repo, orgID)

	// then
	assert.ElementsMatch(t, userIDs, found.Entries())
}

func Test_ActiveUserListRepository_FindByOrganizationID_shouldReturnVersionZero_whenRootRowIsMissing(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID, _ := givenLegacyActiveUsers(ctx, t, tx, "activeuser-legacy-version-org")
	repo := gateway.NewActiveUserListRepository(tx)

	// when
	found := findActiveUserList(ctx, t, repo, orgID)

	// then
	assert.Equal(t, 0, found.Version())
}

func Test_ActiveUserListRepository_Save_shouldCreateRootRow_whenRootRowIsMissing(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID, userIDs := givenLegacyActiveUsers(ctx, t, tx, "activeuser-legacy-save-org")
	repo := gateway.NewActiveUserListRepository(tx)
	list := findActiveUserList(ctx, t, repo, orgID)
	list.Remove(userIDs[0])

	// when
	err := repo.Save(ctx, list)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, findActiveUserList(ctx, t, repo, orgID).Version())
}

func Test_ActiveUserListRepository_Save_shouldKeepEntriesOfBoth_whenLoserRetriesAfterConflict(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "activeuser-both-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "activeuser-both-org", 3)
	repo := gateway.NewActiveUserListRepository(tx)
	saveActiveUsers(ctx, t, repo, orgID, userIDs[0])
	first := findActiveUserList(ctx, t, repo, orgID)
	second := findActiveUserList(ctx, t, repo, orgID)
	require.NoError(t, first.Add(userIDs[1], activeListLimit))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(userIDs[2], activeListLimit))
	require.ErrorIs(t, repo.Save(ctx, second), libversioned.ErrConcurrentModification)
	reread := findActiveUserList(ctx, t, repo, orgID)
	require.NoError(t, reread.Add(userIDs[2], activeListLimit))

	// when
	err := repo.Save(ctx, reread)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, userIDs, findActiveUserList(ctx, t, repo, orgID).Entries())
}
