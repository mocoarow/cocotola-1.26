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
	org := domain.ReconstructOrganization(domain.OrganizationID{}, name, 100, 50)
	require.NoError(t, orgRepo.Save(ctx, org))
	var inserted gateway.OrganizationRecordForTest
	require.NoError(t, tx.Where("name = ?", name).First(&inserted).Error)
	orgID, err := domain.ParseOrganizationID(inserted.ID)
	require.NoError(t, err)
	return orgID
}

func setupUsers(ctx context.Context, t *testing.T, tx *gorm.DB, orgID domain.OrganizationID, orgName string, count int) []domain.AppUserID {
	t.Helper()
	userRepo := gateway.NewAppUserRepository(tx)
	userIDs := make([]domain.AppUserID, count)

	for i := range count {
		loginID := domain.LoginID(fmt.Sprintf("%s-user-%d", orgName, i))
		user := domainuser.ReconstructAppUser(domain.AppUserID{}, orgID, loginID, "", "", "", true)
		require.NoError(t, userRepo.Save(ctx, user))
		var userRec gateway.AppUserRecordForTest
		require.NoError(t, tx.Where("login_id = ?", string(loginID)).First(&userRec).Error)
		uid, err := domain.ParseAppUserID(userRec.ID)
		require.NoError(t, err)
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
		group := domaingroup.ReconstructGroup(domain.GroupID{}, orgID, name, true)
		require.NoError(t, groupRepo.Save(ctx, group))
		var groupRec gateway.GroupRecordForTest
		require.NoError(t, tx.Table("\"group\"").Where("name = ? AND organization_id = ?", name, orgID.String()).First(&groupRec).Error)
		groupIDs[i] = domain.MustParseGroupID(groupRec.ID)
	}
	return groupIDs
}
