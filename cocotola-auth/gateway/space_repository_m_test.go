//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

func setupSpace(ctx context.Context, t *testing.T, repo *gateway.SpaceRepository, orgID domain.OrganizationID, ownerID domain.AppUserID, keyName string, spaceType domainspace.Type, deleted bool) domain.SpaceID {
	t.Helper()
	id, err := domain.NewSpaceIDV7()
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, domainspace.ReconstructSpace(id, orgID, ownerID, keyName, "name-"+keyName, spaceType, deleted)))
	return id
}

func spaceIDs(spaces []domainspace.Space) []domain.SpaceID {
	ids := make([]domain.SpaceID, len(spaces))
	for i := range spaces {
		ids[i] = spaces[i].ID()
	}
	return ids
}

func Test_SpaceRepository_Save_shouldReturnErrDuplicateEntry_whenKeyNameAlreadyExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-duplicate-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-duplicate-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	firstID, err := domain.NewSpaceIDV7()
	require.NoError(t, err)
	secondID, err := domain.NewSpaceIDV7()
	require.NoError(t, err)
	keyName := domainspace.PublicSpaceKeyName("space-duplicate-org")
	require.NoError(t, repo.Save(ctx, domainspace.ReconstructSpace(firstID, orgID, ownerID, keyName, "Public", domainspace.TypePublic(), false)))
	duplicate := domainspace.ReconstructSpace(secondID, orgID, ownerID, keyName, "Public", domainspace.TypePublic(), false)

	// when
	err = repo.Save(ctx, duplicate)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_SpaceRepository_Save_shouldUpdateSpace_whenLoadedAggregateIsModified(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-update-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-update-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	id := setupSpace(ctx, t, repo, orgID, ownerID, "update", domainspace.TypePrivate(), false)
	loaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)

	// when
	loaded.Delete()
	err = repo.Save(ctx, loaded)

	// then
	require.NoError(t, err)
	reloaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.True(t, reloaded.Deleted())
}

func Test_SpaceRepository_Save_shouldReturnConcurrentModification_whenVersionMismatches(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-cas-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-cas-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	id := setupSpace(ctx, t, repo, orgID, ownerID, "cas", domainspace.TypePrivate(), false)
	firstCopy, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	secondCopy, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	firstCopy.Delete()
	require.NoError(t, repo.Save(ctx, firstCopy))

	// when
	err = repo.Save(ctx, secondCopy)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_SpaceRepository_Save_shouldReturnErrSpaceNotFound_whenRowWasDeletedAfterLoad(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-gone-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-gone-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	id := setupSpace(ctx, t, repo, orgID, ownerID, "gone", domainspace.TypePrivate(), false)
	loaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.NoError(t, tx.Exec("DELETE FROM space WHERE id = ?", id.String()).Error)

	// when
	loaded.Delete()
	err = repo.Save(ctx, loaded)

	// then
	require.ErrorIs(t, err, domain.ErrSpaceNotFound)
}

func Test_SpaceRepository_FindByID_shouldReturnSpace_whenSpaceExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-findbyid-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-findbyid-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	id, err := domain.NewSpaceIDV7()
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, domainspace.ReconstructSpace(id, orgID, ownerID, "findbyid-key", "Find By ID", domainspace.TypePrivate(), true)))

	// when
	found, err := repo.FindByID(ctx, id)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, found.ID())
	assert.True(t, orgID.Equal(found.OrganizationID()))
	assert.True(t, ownerID.Equal(found.OwnerID()))
	assert.Equal(t, "findbyid-key", found.KeyName())
	assert.Equal(t, "Find By ID", found.Name())
	assert.Equal(t, domainspace.TypePrivate(), found.SpaceType())
	assert.True(t, found.Deleted())
	assert.Equal(t, 1, found.Version())
}

func Test_SpaceRepository_FindByID_shouldReturnErrSpaceNotFound_whenSpaceDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewSpaceRepository(tx)
	id, err := domain.NewSpaceIDV7()
	require.NoError(t, err)

	// when
	_, err = repo.FindByID(ctx, id)

	// then
	require.ErrorIs(t, err, domain.ErrSpaceNotFound)
}

func Test_SpaceRepository_FindByKeyName_shouldReturnSpaceOfOrganization_whenSameKeyNameExistsInOtherOrganization(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	otherOrgID := setupOrganization(ctx, t, tx, "space-findbykey-other-org")
	otherOwnerID := setupUsers(ctx, t, tx, otherOrgID, "space-findbykey-other-org", 1)[0]
	orgID := setupOrganization(ctx, t, tx, "space-findbykey-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-findbykey-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	setupSpace(ctx, t, repo, otherOrgID, otherOwnerID, "shared-key", domainspace.TypePrivate(), false)
	id := setupSpace(ctx, t, repo, orgID, ownerID, "shared-key", domainspace.TypePrivate(), false)

	// when
	found, err := repo.FindByKeyName(ctx, orgID, "shared-key")

	// then
	require.NoError(t, err)
	assert.Equal(t, id, found.ID())
}

func Test_SpaceRepository_FindByKeyName_shouldReturnErrSpaceNotFound_whenKeyNameDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-findbykey-missing-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-findbykey-missing-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	setupSpace(ctx, t, repo, orgID, ownerID, "existing-key", domainspace.TypePrivate(), false)

	// when
	_, err := repo.FindByKeyName(ctx, orgID, "missing-key")

	// then
	require.ErrorIs(t, err, domain.ErrSpaceNotFound)
}

func Test_SpaceRepository_FindPublicByOrganizationID_shouldReturnPublicSpace_whenPrivateSpaceAlsoExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-public-private-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-public-private-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	setupSpace(ctx, t, repo, orgID, ownerID, "private", domainspace.TypePrivate(), false)
	publicID := setupSpace(ctx, t, repo, orgID, ownerID, "public", domainspace.TypePublic(), false)

	// when
	found, err := repo.FindPublicByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, publicID, found.ID())
}

func Test_SpaceRepository_FindPublicByOrganizationID_shouldReturnPublicSpaceOfOrganization_whenOtherOrganizationHasPublicSpace(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	otherOrgID := setupOrganization(ctx, t, tx, "space-public-other-org")
	otherOwnerID := setupUsers(ctx, t, tx, otherOrgID, "space-public-other-org", 1)[0]
	orgID := setupOrganization(ctx, t, tx, "space-public-own-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-public-own-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	setupSpace(ctx, t, repo, otherOrgID, otherOwnerID, "public", domainspace.TypePublic(), false)
	publicID := setupSpace(ctx, t, repo, orgID, ownerID, "public", domainspace.TypePublic(), false)

	// when
	found, err := repo.FindPublicByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, publicID, found.ID())
}

func Test_SpaceRepository_FindPublicByOrganizationID_shouldReturnErrSpaceNotFound_whenPublicSpaceIsDeleted(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-public-deleted-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-public-deleted-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	setupSpace(ctx, t, repo, orgID, ownerID, "public", domainspace.TypePublic(), true)

	// when
	_, err := repo.FindPublicByOrganizationID(ctx, orgID)

	// then
	require.ErrorIs(t, err, domain.ErrSpaceNotFound)
}

func Test_SpaceRepository_FindByOrganizationID_shouldReturnSpacesOfOrganization_whenOtherOrganizationHasSpaces(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	otherOrgID := setupOrganization(ctx, t, tx, "space-list-other-org")
	otherOwnerID := setupUsers(ctx, t, tx, otherOrgID, "space-list-other-org", 1)[0]
	orgID := setupOrganization(ctx, t, tx, "space-list-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-list-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	setupSpace(ctx, t, repo, otherOrgID, otherOwnerID, "public", domainspace.TypePublic(), false)
	publicID := setupSpace(ctx, t, repo, orgID, ownerID, "public", domainspace.TypePublic(), false)
	privateID := setupSpace(ctx, t, repo, orgID, ownerID, "private", domainspace.TypePrivate(), false)

	// when
	spaces, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.SpaceID{publicID, privateID}, spaceIDs(spaces))
}

func Test_SpaceRepository_FindByOrganizationID_shouldExcludeDeletedSpaces_whenDeletedSpaceExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "space-list-deleted-org")
	ownerID := setupUsers(ctx, t, tx, orgID, "space-list-deleted-org", 1)[0]
	repo := gateway.NewSpaceRepository(tx)
	activeID := setupSpace(ctx, t, repo, orgID, ownerID, "active", domainspace.TypePrivate(), false)
	setupSpace(ctx, t, repo, orgID, ownerID, "deleted", domainspace.TypePrivate(), true)

	// when
	spaces, err := repo.FindByOrganizationID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, []domain.SpaceID{activeID}, spaceIDs(spaces))
}
