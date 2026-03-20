//go:build medium

package gateway_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func randOrgID(t *testing.T) int {
	t.Helper()
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	require.NoError(t, err)
	return int(n.Int64()) + 1
}

func mustRole(t *testing.T, name string) domain.RBACRole {
	t.Helper()
	r, err := domain.NewRBACRole(name)
	require.NoError(t, err)
	return r
}

func mustResource(t *testing.T, name string) domain.RBACResource {
	t.Helper()
	r, err := domain.NewRBACResource(name)
	require.NoError(t, err)
	return r
}

func Test_RBACRepository_AddPolicy_shouldEnforceDirectPolicy_whenUserHasNoRole(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	aliceID := 100
	bobID := 200

	aliceRole := mustRole(t, fmt.Sprintf("org:%d,alice_role", orgID))
	data1 := mustResource(t, fmt.Sprintf("org:%d,data:1", orgID))
	data2 := mustResource(t, fmt.Sprintf("org:%d,data:2", orgID))

	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, aliceID, aliceRole))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, aliceRole, domain.ActionViewUser(), data1, domain.EffectAllow()))

	bobRole := mustRole(t, fmt.Sprintf("org:%d,bob_role", orgID))
	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, bobID, bobRole))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, bobRole, domain.ActionCreateUser(), data2, domain.EffectAllow()))

	tests := []struct {
		userID   int
		action   domain.RBACAction
		resource domain.RBACResource
		want     bool
	}{
		{userID: aliceID, action: domain.ActionViewUser(), resource: data1, want: true},
		{userID: aliceID, action: domain.ActionCreateUser(), resource: data1, want: false},
		{userID: aliceID, action: domain.ActionViewUser(), resource: data2, want: false},
		{userID: bobID, action: domain.ActionCreateUser(), resource: data2, want: true},
		{userID: bobID, action: domain.ActionViewUser(), resource: data2, want: false},
		{userID: bobID, action: domain.ActionCreateUser(), resource: data1, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user:%d,%s,%s,want:%v", tt.userID, tt.action.Value(), tt.resource.Value(), tt.want), func(t *testing.T) {
			t.Parallel()
			// when
			ok, err := rbacRepo.Enforce(orgID, tt.userID, tt.action, tt.resource)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.want, ok)
		})
	}
}

func Test_RBACRepository_AssignRoleToUser_shouldInheritPolicies_whenRoleHasPolicies(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	aliceID := 100
	bobID := 200

	readerRole := mustRole(t, fmt.Sprintf("org:%d,reader", orgID))
	writerRole := mustRole(t, fmt.Sprintf("org:%d,writer", orgID))
	data1 := mustResource(t, fmt.Sprintf("org:%d,data:1", orgID))
	data2 := mustResource(t, fmt.Sprintf("org:%d,data:2", orgID))

	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerRole, domain.ActionViewUser(), data1, domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, writerRole, domain.ActionCreateUser(), data2, domain.EffectAllow()))

	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, aliceID, readerRole))
	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, bobID, writerRole))

	tests := []struct {
		userID   int
		action   domain.RBACAction
		resource domain.RBACResource
		want     bool
	}{
		{userID: aliceID, action: domain.ActionViewUser(), resource: data1, want: true},
		{userID: aliceID, action: domain.ActionCreateUser(), resource: data1, want: false},
		{userID: aliceID, action: domain.ActionViewUser(), resource: data2, want: false},
		{userID: bobID, action: domain.ActionCreateUser(), resource: data2, want: true},
		{userID: bobID, action: domain.ActionViewUser(), resource: data1, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user:%d,%s,%s,want:%v", tt.userID, tt.action.Value(), tt.resource.Value(), tt.want), func(t *testing.T) {
			t.Parallel()
			ok, err := rbacRepo.Enforce(orgID, tt.userID, tt.action, tt.resource)
			require.NoError(t, err)
			assert.Equal(t, tt.want, ok)
		})
	}
}

func Test_RBACRepository_AddObjectGroupingPolicy_shouldInheritAccess_whenResourceHasParent(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	aliceID := 100

	readerRole := mustRole(t, fmt.Sprintf("org:%d,reader", orgID))
	data1 := mustResource(t, fmt.Sprintf("org:%d,data:1", orgID))
	child1 := mustResource(t, fmt.Sprintf("org:%d,child:1", orgID))

	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, aliceID, readerRole))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerRole, domain.ActionViewUser(), data1, domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddObjectGroupingPolicy(ctx, orgID, child1, data1))

	tests := []struct {
		resource domain.RBACResource
		want     bool
	}{
		{resource: data1, want: true},
		{resource: child1, want: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s,want:%v", tt.resource.Value(), tt.want), func(t *testing.T) {
			t.Parallel()
			ok, err := rbacRepo.Enforce(orgID, aliceID, domain.ActionViewUser(), tt.resource)
			require.NoError(t, err)
			assert.Equal(t, tt.want, ok)
		})
	}
}

func Test_RBACRepository_AddPolicy_shouldDenyOverrideAllow_whenBothExist(t *testing.T) {
	t.Parallel()
	// given
	// hierarchy: data:1 / data:2 / data:3 / data:4 / data:5
	// allow on data:2, deny on data:4
	// result: data:1=no, data:2=yes, data:3=yes, data:4=no, data:5=no
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	aliceID := 100
	readerRole := mustRole(t, fmt.Sprintf("org:%d,reader", orgID))

	resources := make([]domain.RBACResource, 5)
	for i := range 5 {
		resources[i] = mustResource(t, fmt.Sprintf("org:%d,data:%d", orgID, i+1))
	}

	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, aliceID, readerRole))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerRole, domain.ActionViewUser(), resources[1], domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerRole, domain.ActionViewUser(), resources[3], domain.EffectDeny()))

	// Build hierarchy: data:2 is child of data:1, data:3 of data:2, etc.
	for i := 1; i < 5; i++ {
		require.NoError(t, rbacRepo.AddObjectGroupingPolicy(ctx, orgID, resources[i], resources[i-1]))
	}

	tests := []struct {
		resource domain.RBACResource
		want     bool
	}{
		{resource: resources[0], want: false},
		{resource: resources[1], want: true},
		{resource: resources[2], want: true},
		{resource: resources[3], want: false},
		{resource: resources[4], want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s,want:%v", tt.resource.Value(), tt.want), func(t *testing.T) {
			t.Parallel()
			ok, err := rbacRepo.Enforce(orgID, aliceID, domain.ActionViewUser(), tt.resource)
			require.NoError(t, err)
			assert.Equal(t, tt.want, ok)
		})
	}
}

func Test_CasbinAuthorizationChecker_IsAllowed_shouldReturnTrue_whenUserHasPermission(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	checker := gateway.NewCasbinAuthorizationChecker(rbacRepo)

	orgID := randOrgID(t)
	userID := 100
	adminRole := mustRole(t, fmt.Sprintf("org:%d,admin", orgID))

	require.NoError(t, rbacRepo.AssignRoleToUser(ctx, orgID, userID, adminRole))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, adminRole, domain.ActionCreateUser(), domain.ResourceAny(), domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, adminRole, domain.ActionViewUser(), domain.ResourceAny(), domain.EffectAllow()))

	// when
	ok, err := checker.IsAllowed(ctx, orgID, userID, domain.ActionCreateUser(), domain.ResourceAny())

	// then
	require.NoError(t, err)
	assert.True(t, ok)
}

func Test_CasbinAuthorizationChecker_IsAllowed_shouldReturnFalse_whenUserLacksPermission(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	checker := gateway.NewCasbinAuthorizationChecker(rbacRepo)

	orgID := randOrgID(t)
	userID := 100

	// when (no role assigned)
	ok, err := checker.IsAllowed(ctx, orgID, userID, domain.ActionCreateUser(), domain.ResourceAny())

	// then
	require.NoError(t, err)
	assert.False(t, ok)
}
