//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
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
	user := domainuser.ReconstructAppUser(0, orgID, "testuser@example.com", "", "", "", true)

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
	user := domainuser.ReconstructAppUser(0, orgID, "findbyid@example.com", "", "", "", true)
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
	user := domainuser.ReconstructAppUser(0, orgID, "login@example.com", "", "", "", true)
	require.NoError(t, repo.Save(ctx, user))

	// when
	found, err := repo.FindByLoginID(ctx, orgID, domain.LoginID("login@example.com"))

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
	_, err := repo.FindByLoginID(ctx, 1, domain.LoginID("nonexistent@example.com"))

	// then
	require.ErrorIs(t, err, domain.ErrAppUserNotFound)
}

func Test_AppUserRepository_Save_shouldPersistHashedPassword_whenDomainHasPassword(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-pw-org")
	hashedPw := "$2a$10$abcdefghij"

	repo := gateway.NewAppUserRepository(tx)
	user := domainuser.ReconstructAppUser(0, orgID, "pw-test@example.com", hashedPw, "", "", true)
	require.NoError(t, repo.Save(ctx, user))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "pw-test@example.com").First(&inserted).Error)

	// when
	// Load the persisted aggregate via FindByID so its version is populated from the
	// existing row — otherwise the version-0 Save path would attempt another INSERT.
	loaded, err := repo.FindByID(ctx, inserted.ID)
	require.NoError(t, err)
	newHashedPw := "$2a$10$newhashedpw"
	updated := domainuser.
		ReconstructAppUser(loaded.ID(), loaded.OrganizationID(), loaded.LoginID(), newHashedPw, loaded.Provider(), loaded.ProviderID(), false).
		WithVersion(loaded.Version())
	err = repo.Save(ctx, updated)

	// then
	require.NoError(t, err)
	var afterUpdate gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("id = ?", inserted.ID).First(&afterUpdate).Error)
	assert.False(t, afterUpdate.Enabled)
	require.NotNil(t, afterUpdate.HashedPassword)
	assert.Equal(t, newHashedPw, *afterUpdate.HashedPassword)
}

func Test_AppUserRepository_Save_shouldReturnConcurrentModification_whenVersionMismatches(t *testing.T) {
	t.Parallel()
	// given: persist a user and load two independent in-memory copies of the same
	// aggregate, simulating two concurrent transactions each holding its own snapshot.
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-cas-org")
	repo := gateway.NewAppUserRepository(tx)

	initial := domainuser.ReconstructAppUser(0, orgID, "cas@example.com", "", "", "", true)
	require.NoError(t, repo.Save(ctx, initial))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "cas@example.com").First(&inserted).Error)

	firstCopy, err := repo.FindByID(ctx, inserted.ID)
	require.NoError(t, err)
	secondCopy, err := repo.FindByID(ctx, inserted.ID)
	require.NoError(t, err)

	// when: the first transaction commits its update (bumping the stored version),
	// then the second transaction — still holding the stale version — tries to save.
	firstCopy.Disable()
	require.NoError(t, repo.Save(ctx, firstCopy))

	secondCopy.Enable()
	err = repo.Save(ctx, secondCopy)

	// then: the CAS must fail and the caller must be told to reload.
	require.ErrorIs(t, err, domain.ErrAppUserConcurrentModification)

	// And the row must reflect the first commit, not the stale second save.
	var afterUpdate gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("id = ?", inserted.ID).First(&afterUpdate).Error)
	assert.False(t, afterUpdate.Enabled)
}
