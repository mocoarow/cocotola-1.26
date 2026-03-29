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

func setupOrganization(ctx context.Context, t *testing.T, tx *gorm.DB, name string) int {
	t.Helper()
	orgRepo := gateway.NewOrganizationRepository(tx)
	org := domain.ReconstructOrganization(0, name, 100, 50)
	require.NoError(t, orgRepo.Save(ctx, org))
	var inserted gateway.OrganizationRecordForTest
	require.NoError(t, tx.Where("name = ?", name).First(&inserted).Error)
	return inserted.ID
}

func setupUsers(ctx context.Context, t *testing.T, tx *gorm.DB, orgID int, orgName string, count int) []int {
	t.Helper()
	userRepo := gateway.NewAppUserRepository(tx)
	userIDs := make([]int, count)

	for i := range count {
		loginID := domain.LoginID(fmt.Sprintf("%s-user-%d", orgName, i))
		user := domainuser.ReconstructAppUser(0, orgID, loginID, "", true)
		require.NoError(t, userRepo.Save(ctx, user))
		var userRec gateway.AppUserRecordForTest
		require.NoError(t, tx.Where("login_id = ?", string(loginID)).First(&userRec).Error)
		userIDs[i] = userRec.ID
	}
	return userIDs
}

func setupGroups(ctx context.Context, t *testing.T, tx *gorm.DB, orgID int, orgName string, count int) []int {
	t.Helper()
	groupRepo := gateway.NewGroupRepository(tx)
	groupIDs := make([]int, count)

	for i := range count {
		name := fmt.Sprintf("%s-group-%d", orgName, i)
		group := domaingroup.ReconstructGroup(0, orgID, name, true)
		require.NoError(t, groupRepo.Save(ctx, group))
		var groupRec gateway.GroupRecordForTest
		require.NoError(t, tx.Table("\"group\"").Where("name = ? AND organization_id = ?", name, orgID).First(&groupRec).Error)
		groupIDs[i] = groupRec.ID
	}
	return groupIDs
}
