//go:build medium

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func Test_AppUserProviderRepository_Save_shouldReturnErrDuplicateEntry_whenProviderIDAlreadyLinked(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	orgID := setupOrganization(ctx, t, tx, "provider-duplicate-org")
	userIDs := setupUsers(ctx, t, tx, orgID, "provider-duplicate-org", 2)
	repo := gateway.NewAppUserProviderRepository(tx)
	firstID, err := domain.NewAppUserProviderIDV7()
	require.NoError(t, err)
	secondID, err := domain.NewAppUserProviderIDV7()
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, domainuser.ReconstructAppUserProvider(firstID, userIDs[0], orgID, "supabase", "sub-duplicate")))
	duplicate := domainuser.ReconstructAppUserProvider(secondID, userIDs[1], orgID, "supabase", "sub-duplicate")

	// when
	err = repo.Save(ctx, duplicate)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}
