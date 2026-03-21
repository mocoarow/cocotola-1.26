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

func mustGroup(t *testing.T, name string) domain.RBACGroup {
	t.Helper()
	g, err := domain.NewRBACGroup(name)
	require.NoError(t, err)
	return g
}

func mustResource(t *testing.T, name string) domain.RBACResource {
	t.Helper()
	r, err := domain.NewRBACResource(name)
	require.NoError(t, err)
	return r
}

func Test_RBACRepository_AddPolicy_shouldEnforceDirectPolicy_whenUserHasNoGroup(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	aliceID := 100
	bobID := 200

	aliceGroup := mustGroup(t, fmt.Sprintf("org:%d,alice_group", orgID))
	data1 := mustResource(t, fmt.Sprintf("org:%d,data:1", orgID))
	data2 := mustResource(t, fmt.Sprintf("org:%d,data:2", orgID))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, aliceGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, aliceGroup, domain.ActionViewUser(), data1, domain.EffectAllow()))

	bobGroup := mustGroup(t, fmt.Sprintf("org:%d,bob_group", orgID))
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, bobID, bobGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, bobGroup, domain.ActionCreateUser(), data2, domain.EffectAllow()))

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

func Test_RBACRepository_AssignGroupToUser_shouldInheritPolicies_whenGroupHasPolicies(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	aliceID := 100
	bobID := 200

	readerGroup := mustGroup(t, fmt.Sprintf("org:%d,reader", orgID))
	writerGroup := mustGroup(t, fmt.Sprintf("org:%d,writer", orgID))
	data1 := mustResource(t, fmt.Sprintf("org:%d,data:1", orgID))
	data2 := mustResource(t, fmt.Sprintf("org:%d,data:2", orgID))

	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domain.ActionViewUser(), data1, domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, writerGroup, domain.ActionCreateUser(), data2, domain.EffectAllow()))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, readerGroup))
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, bobID, writerGroup))

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

	readerGroup := mustGroup(t, fmt.Sprintf("org:%d,reader", orgID))
	data1 := mustResource(t, fmt.Sprintf("org:%d,data:1", orgID))
	child1 := mustResource(t, fmt.Sprintf("org:%d,child:1", orgID))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, readerGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domain.ActionViewUser(), data1, domain.EffectAllow()))
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
	readerGroup := mustGroup(t, fmt.Sprintf("org:%d,reader", orgID))

	resources := make([]domain.RBACResource, 5)
	for i := range 5 {
		resources[i] = mustResource(t, fmt.Sprintf("org:%d,data:%d", orgID, i+1))
	}

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, readerGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domain.ActionViewUser(), resources[1], domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domain.ActionViewUser(), resources[3], domain.EffectDeny()))

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
	adminGroup := mustGroup(t, fmt.Sprintf("org:%d,admin", orgID))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, adminGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, adminGroup, domain.ActionCreateUser(), domain.ResourceAny(), domain.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, adminGroup, domain.ActionViewUser(), domain.ResourceAny(), domain.EffectAllow()))

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

	// when (no group assigned)
	ok, err := checker.IsAllowed(ctx, orgID, userID, domain.ActionCreateUser(), domain.ResourceAny())

	// then
	require.NoError(t, err)
	assert.False(t, ok)
}

func Test_RBACRepository_GetGroupsForUser_shouldReturnGroups_whenUserHasGroups(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	userID := 100

	adminGroup := mustGroup(t, fmt.Sprintf("org:%d,admin", orgID))
	editorGroup := mustGroup(t, fmt.Sprintf("org:%d,editor", orgID))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, adminGroup))
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, editorGroup))

	// when
	groups, err := rbacRepo.GetGroupsForUser(ctx, orgID, userID)

	// then
	require.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.Contains(t, groups, adminGroup.Value())
	assert.Contains(t, groups, editorGroup.Value())
}

func Test_RBACRepository_GetGroupsForUser_shouldReturnEmpty_whenUserHasNoGroups(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	userID := 100

	// when
	groups, err := rbacRepo.GetGroupsForUser(ctx, orgID, userID)

	// then
	require.NoError(t, err)
	assert.Empty(t, groups)
}
