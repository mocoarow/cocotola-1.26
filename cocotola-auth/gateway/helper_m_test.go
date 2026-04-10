//go:build medium

package gateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func setupOrganization(ctx context.Context, t *testing.T, tx *gorm.DB, name string) domain.OrganizationID {
	t.Helper()
	orgRepo := gateway.NewOrganizationRepository(tx)
	orgID, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	org := domain.ReconstructOrganization(orgID, name, 100, 50)
	require.NoError(t, orgRepo.Save(ctx, org))
	return orgID
}

func setupUsers(ctx context.Context, t *testing.T, tx *gorm.DB, orgID domain.OrganizationID, orgName string, count int) []domain.AppUserID {
	t.Helper()
	userRepo := gateway.NewAppUserRepository(tx)
	userIDs := make([]domain.AppUserID, count)

	for i := range count {
		loginID := domain.LoginID(fmt.Sprintf("%s-user-%d", orgName, i))
		uid, err := domain.NewAppUserIDV7()
		require.NoError(t, err)
		user := domainuser.ReconstructAppUser(uid, orgID, loginID, "", "", "", true)
		require.NoError(t, userRepo.Save(ctx, user))
		userIDs[i] = uid
	}
	return userIDs
}

func setupGroups(ctx context.Context, t *testing.T, tx *gorm.DB, orgID domain.OrganizationID, orgName string, count int) []domain.GroupID {
	t.Helper()
	groupRepo := gateway.NewGroupRepository(tx)
	groupIDs := make([]domain.GroupID, count)

	for i := range count {
		name := fmt.Sprintf("%s-group-%d", orgName, i)
		gid, err := domain.NewGroupIDV7()
		require.NoError(t, err)
		group := domaingroup.ReconstructGroup(gid, orgID, name, true)
		require.NoError(t, groupRepo.Save(ctx, group))
		groupIDs[i] = gid
	}
	return groupIDs
}
