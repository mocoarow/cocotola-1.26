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

func Test_AppUserRepository_Save_shouldInsertAppUser_whenNewRecord(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-save-org")
	repo := gateway.NewAppUserRepository(tx)
	user := domain.ReconstructAppUser(0, orgID, "testuser@example.com", true)

	// when
	err := repo.Save(ctx, user)

	// then
	require.NoError(t, err)
}

func Test_AppUserRepository_FindByID_shouldReturnAppUser_whenUserExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-findbyid-org")
	repo := gateway.NewAppUserRepository(tx)
	user := domain.ReconstructAppUser(0, orgID, "findbyid@example.com", true)
	require.NoError(t, repo.Save(ctx, user))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "findbyid@example.com").First(&inserted).Error)

	// when
	found, err := repo.FindByID(ctx, inserted.ID)

	// then
	require.NoError(t, err)
	assert.Equal(t, inserted.ID, found.ID())
	assert.Equal(t, orgID, found.OrganizationID())
	assert.Equal(t, domain.LoginID("findbyid@example.com"), found.LoginID())
	assert.True(t, found.Enabled())
}

func Test_AppUserRepository_FindByID_shouldReturnError_whenUserDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewAppUserRepository(tx)

	// when
	_, err := repo.FindByID(ctx, 999999)

	// then
	require.ErrorIs(t, err, domain.ErrAppUserNotFound)
}

func Test_AppUserRepository_FindByLoginID_shouldReturnAppUser_whenLoginIDExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-findbyloginid-org")
	repo := gateway.NewAppUserRepository(tx)
	user := domain.ReconstructAppUser(0, orgID, "login@example.com", true)
	require.NoError(t, repo.Save(ctx, user))

	// when
	found, err := repo.FindByLoginID(ctx, orgID, "login@example.com")

	// then
	require.NoError(t, err)
	assert.Equal(t, orgID, found.OrganizationID())
	assert.Equal(t, domain.LoginID("login@example.com"), found.LoginID())
}

func Test_AppUserRepository_FindByLoginID_shouldReturnError_whenLoginIDDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewAppUserRepository(tx)

	// when
	_, err := repo.FindByLoginID(ctx, 1, "nonexistent@example.com")

	// then
	require.ErrorIs(t, err, domain.ErrAppUserNotFound)
}

func Test_AppUserRepository_Save_shouldNotOverwriteHashedPassword_whenUpdating(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-omit-org")
	hashedPw := "$2a$10$abcdefghij"
	tx.Exec("INSERT INTO app_user (created_by, updated_by, organization_id, login_id, hashed_password, enabled) VALUES (0, 0, ?, 'omit-test@example.com', ?, 1)", orgID, hashedPw)

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "omit-test@example.com").First(&inserted).Error)

	repo := gateway.NewAppUserRepository(tx)
	updated := domain.ReconstructAppUser(inserted.ID, orgID, "omit-test@example.com", false)

	// when
	err := repo.Save(ctx, updated)

	// then
	require.NoError(t, err)
	var afterUpdate gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("id = ?", inserted.ID).First(&afterUpdate).Error)
	assert.False(t, afterUpdate.Enabled)
	assert.NotNil(t, afterUpdate.HashedPassword)
	assert.Equal(t, hashedPw, *afterUpdate.HashedPassword)
}
