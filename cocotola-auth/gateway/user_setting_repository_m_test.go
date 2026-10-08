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

func Test_UserSettingRepository_FindByAppUserID_shouldReturnSystemAdminQuota_whenMigrationsApplied(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	repo := gateway.NewUserSettingRepository(testDB)
	const systemAdminMaxWorkbooks = 100

	// when
	setting, err := repo.FindByAppUserID(ctx, domain.SystemAppUserID())

	// then
	require.NoError(t, err)
	assert.Equal(t, systemAdminMaxWorkbooks, setting.MaxWorkbooks())
}
