//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

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
