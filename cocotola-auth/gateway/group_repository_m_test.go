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
)

func Test_GroupRepository_Save_shouldInsertGroup_whenNewRecord(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "group-save-org")
	repo := gateway.NewGroupRepository(tx)
	group := domaingroup.ReconstructGroup(domain.GroupID{}, orgID, "test-group", true)

	// when
	err := repo.Save(ctx, group)

	// then
	require.NoError(t, err)
}

func Test_GroupRepository_FindByID_shouldReturnGroup_whenGroupExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "group-findbyid-org")
	repo := gateway.NewGroupRepository(tx)
	group := domaingroup.ReconstructGroup(domain.GroupID{}, orgID, "findbyid-group", true)
	require.NoError(t, repo.Save(ctx, group))

	var inserted gateway.GroupRecordForTest
	require.NoError(t, tx.Table("\"group\"").Where("name = ? AND organization_id = ?", "findbyid-group", orgID.String()).First(&inserted).Error)
	insertedID := domain.MustParseGroupID(inserted.ID)

	// when
	found, err := repo.FindByID(ctx, insertedID)

	// then
	require.NoError(t, err)
	assert.Equal(t, insertedID, found.ID())
	assert.True(t, orgID.Equal(found.OrganizationID()))
	assert.Equal(t, "findbyid-group", found.Name())
	assert.True(t, found.Enabled())
}

func Test_GroupRepository_FindByID_shouldReturnError_whenGroupDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewGroupRepository(tx)

	// when
	_, err := repo.FindByID(ctx, domain.MustParseGroupID("00000000-0000-7000-8000-ffffffffffff"))

	// then
	require.ErrorIs(t, err, domain.ErrGroupNotFound)
}

func Test_GroupRepository_FindByName_shouldReturnGroup_whenNameExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "group-findbyname-org")
	repo := gateway.NewGroupRepository(tx)
	group := domaingroup.ReconstructGroup(domain.GroupID{}, orgID, "findbyname-group", true)
	require.NoError(t, repo.Save(ctx, group))

	// when
	found, err := repo.FindByName(ctx, orgID, "findbyname-group")

	// then
	require.NoError(t, err)
	assert.True(t, orgID.Equal(found.OrganizationID()))
	assert.Equal(t, "findbyname-group", found.Name())
}

func Test_GroupRepository_FindByName_shouldReturnError_whenNameDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewGroupRepository(tx)
	nonExistentOrgID := domain.MustParseOrganizationID("00000000-0000-7000-8000-ffffffffffff")

	// when
	_, err := repo.FindByName(ctx, nonExistentOrgID, "nonexistent-group")

	// then
	require.ErrorIs(t, err, domain.ErrGroupNotFound)
}
