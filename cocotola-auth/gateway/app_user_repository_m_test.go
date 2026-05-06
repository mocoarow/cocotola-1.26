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
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

func Test_AppUserRepository_Save_shouldInsertAppUser_whenNewRecord(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-save-org")
	repo := gateway.NewAppUserRepository(tx)
	user := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, "testuser@example.com", "", true)

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
	user := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, "findbyid@example.com", "", true)
	require.NoError(t, repo.Save(ctx, user))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "findbyid@example.com").First(&inserted).Error)
	insertedID, err := domain.ParseAppUserID(inserted.ID)
	require.NoError(t, err)

	// when
	found, err := repo.FindByID(ctx, insertedID)

	// then
	require.NoError(t, err)
	assert.True(t, insertedID.Equal(found.ID()))
	assert.True(t, orgID.Equal(found.OrganizationID()))
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
	nonExistentID := domain.MustParseAppUserID("00000000-0000-7000-8000-ffffffffffff")

	// when
	_, err := repo.FindByID(ctx, nonExistentID)

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
	user := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, "login@example.com", "", true)
	require.NoError(t, repo.Save(ctx, user))

	// when
	found, err := repo.FindByLoginID(ctx, orgID, domain.LoginID("login@example.com"))

	// then
	require.NoError(t, err)
	assert.True(t, orgID.Equal(found.OrganizationID()))
	assert.Equal(t, domain.LoginID("login@example.com"), found.LoginID())
}

func Test_AppUserRepository_FindByLoginID_shouldReturnError_whenLoginIDDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewAppUserRepository(tx)
	nonExistentOrgID := domain.MustParseOrganizationID("00000000-0000-7000-8000-ffffffffffff")

	// when
	_, err := repo.FindByLoginID(ctx, nonExistentOrgID, domain.LoginID("nonexistent@example.com"))

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
	user := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, "pw-test@example.com", hashedPw, true)
	require.NoError(t, repo.Save(ctx, user))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "pw-test@example.com").First(&inserted).Error)
	insertedID, err := domain.ParseAppUserID(inserted.ID)
	require.NoError(t, err)

	// when
	// Load the persisted aggregate via FindByID so its version is populated from the
	// existing row — otherwise the version-0 Save path would attempt another INSERT.
	loaded, err := repo.FindByID(ctx, insertedID)
	require.NoError(t, err)
	newHashedPw := "$2a$10$newhashedpw"
	updated := domainuser.ReconstructAppUser(loaded.ID(), loaded.OrganizationID(), loaded.LoginID(), newHashedPw, false)
	updated.SetVersion(loaded.Version())
	err = repo.Save(ctx, updated)

	// then
	require.NoError(t, err)
	var afterUpdate gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("id = ?", inserted.ID).First(&afterUpdate).Error)
	assert.False(t, afterUpdate.Enabled)
	require.NotNil(t, afterUpdate.HashedPassword)
	assert.Equal(t, newHashedPw, *afterUpdate.HashedPassword)
}

func Test_AppUserRepository_Save_shouldReturnErrAppUserNotFound_whenRowWasDeletedAfterLoad(t *testing.T) {
	t.Parallel()
	// given: a saved app user is deleted from storage while a stale aggregate is held in memory
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "appuser-deleted-org")
	repo := gateway.NewAppUserRepository(tx)

	initial := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, "deleted@example.com", "", true)
	require.NoError(t, repo.Save(ctx, initial))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "deleted@example.com").First(&inserted).Error)
	insertedID, err := domain.ParseAppUserID(inserted.ID)
	require.NoError(t, err)

	loaded, err := repo.FindByID(ctx, insertedID)
	require.NoError(t, err)

	// delete the underlying row out-of-band (raw SQL to bypass the repository)
	require.NoError(t, tx.Exec("DELETE FROM app_user WHERE id = ?", inserted.ID).Error)

	// when: the stale loaded aggregate tries to save
	loaded.Disable()
	err = repo.Save(ctx, loaded)

	// then: callers see a domain not-found, not a generic error
	require.ErrorIs(t, err, domain.ErrAppUserNotFound)
	assert.NotErrorIs(t, err, libversioned.ErrConcurrentModification,
		"deleted row must surface as not-found, not as concurrent modification")
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

	initial := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, "cas@example.com", "", true)
	require.NoError(t, repo.Save(ctx, initial))

	var inserted gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("login_id = ?", "cas@example.com").First(&inserted).Error)
	insertedID, err := domain.ParseAppUserID(inserted.ID)
	require.NoError(t, err)

	firstCopy, err := repo.FindByID(ctx, insertedID)
	require.NoError(t, err)
	secondCopy, err := repo.FindByID(ctx, insertedID)
	require.NoError(t, err)

	// when: the first transaction commits its update (bumping the stored version),
	// then the second transaction — still holding the stale version — tries to save.
	firstCopy.Disable()
	require.NoError(t, repo.Save(ctx, firstCopy))

	secondCopy.Enable()
	err = repo.Save(ctx, secondCopy)

	// then: the CAS must fail and the caller must be told to reload.
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)

	// And the row must reflect the first commit, not the stale second save.
	var afterUpdate gateway.AppUserRecordForTest
	require.NoError(t, tx.Where("id = ?", inserted.ID).First(&afterUpdate).Error)
	assert.False(t, afterUpdate.Enabled)
}
