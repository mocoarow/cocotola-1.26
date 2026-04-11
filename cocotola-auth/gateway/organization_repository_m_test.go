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

func Test_OrganizationRepository_Save_shouldInsertOrganization_whenNewRecord(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewOrganizationRepository(tx)
	orgID, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	org, err := domain.NewOrganization(orgID, "test-org-save", 10, 5)
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, org)

	// then
	require.NoError(t, err)
}

func Test_OrganizationRepository_FindByID_shouldReturnOrganization_whenOrganizationExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewOrganizationRepository(tx)
	orgID, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	org, err := domain.NewOrganization(orgID, "test-org-findbyid", 20, 10)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, org))

	// when
	found, err := repo.FindByID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.True(t, orgID.Equal(found.ID()))
	assert.Equal(t, "test-org-findbyid", found.Name())
	assert.Equal(t, 20, found.MaxActiveUsers())
	assert.Equal(t, 10, found.MaxActiveGroups())
}

func Test_OrganizationRepository_FindByID_shouldReturnError_whenOrganizationDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewOrganizationRepository(tx)
	nonExistentID := domain.MustParseOrganizationID("00000000-0000-7000-8000-ffffffffffff")

	// when
	_, err := repo.FindByID(ctx, nonExistentID)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
}

func Test_OrganizationRepository_FindByName_shouldReturnOrganization_whenNameExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewOrganizationRepository(tx)
	orgID, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	org, err := domain.NewOrganization(orgID, "test-org-findbyname", 30, 15)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, org))

	// when
	found, err := repo.FindByName(ctx, "test-org-findbyname")

	// then
	require.NoError(t, err)
	assert.Equal(t, "test-org-findbyname", found.Name())
	assert.Equal(t, 30, found.MaxActiveUsers())
	assert.Equal(t, 15, found.MaxActiveGroups())
}

func Test_OrganizationRepository_FindByName_shouldReturnError_whenNameDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewOrganizationRepository(tx)

	// when
	_, err := repo.FindByName(ctx, "nonexistent-org")

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
}
