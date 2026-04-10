//go:build medium

package gateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func randOrgID(t *testing.T) domain.OrganizationID {
	t.Helper()
	id, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	return id
}

func randAppUserID(t *testing.T) domain.AppUserID {
	t.Helper()
	id, err := domain.NewAppUserIDV7()
	require.NoError(t, err)
	return id
}

func mustGroup(t *testing.T, name string) domainrbac.Group {
	t.Helper()
	g, err := domainrbac.NewGroup(name)
	require.NoError(t, err)
	return g
}

func mustResource(t *testing.T, name string) domainrbac.Resource {
	t.Helper()
	r, err := domainrbac.NewResource(name)
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
	aliceID := randAppUserID(t)
	bobID := randAppUserID(t)

	aliceGroup := mustGroup(t, fmt.Sprintf("org:%s,alice_group", orgID.String()))
	data1 := mustResource(t, fmt.Sprintf("org:%s,data:1", orgID.String()))
	data2 := mustResource(t, fmt.Sprintf("org:%s,data:2", orgID.String()))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, aliceGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, aliceGroup, domainrbac.ActionViewUser(), data1, domainrbac.EffectAllow()))

	bobGroup := mustGroup(t, fmt.Sprintf("org:%s,bob_group", orgID.String()))
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, bobID, bobGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, bobGroup, domainrbac.ActionCreateUser(), data2, domainrbac.EffectAllow()))

	tests := []struct {
		userID   domain.AppUserID
		action   domainrbac.Action
		resource domainrbac.Resource
		want     bool
	}{
		{userID: aliceID, action: domainrbac.ActionViewUser(), resource: data1, want: true},
		{userID: aliceID, action: domainrbac.ActionCreateUser(), resource: data1, want: false},
		{userID: aliceID, action: domainrbac.ActionViewUser(), resource: data2, want: false},
		{userID: bobID, action: domainrbac.ActionCreateUser(), resource: data2, want: true},
		{userID: bobID, action: domainrbac.ActionViewUser(), resource: data2, want: false},
		{userID: bobID, action: domainrbac.ActionCreateUser(), resource: data1, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user:%s,%s,%s,want:%v", tt.userID.String(), tt.action.Value(), tt.resource.Value(), tt.want), func(t *testing.T) {
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
	aliceID := randAppUserID(t)
	bobID := randAppUserID(t)

	readerGroup := mustGroup(t, fmt.Sprintf("org:%s,reader", orgID.String()))
	writerGroup := mustGroup(t, fmt.Sprintf("org:%s,writer", orgID.String()))
	data1 := mustResource(t, fmt.Sprintf("org:%s,data:1", orgID.String()))
	data2 := mustResource(t, fmt.Sprintf("org:%s,data:2", orgID.String()))

	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domainrbac.ActionViewUser(), data1, domainrbac.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, writerGroup, domainrbac.ActionCreateUser(), data2, domainrbac.EffectAllow()))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, readerGroup))
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, bobID, writerGroup))

	tests := []struct {
		userID   domain.AppUserID
		action   domainrbac.Action
		resource domainrbac.Resource
		want     bool
	}{
		{userID: aliceID, action: domainrbac.ActionViewUser(), resource: data1, want: true},
		{userID: aliceID, action: domainrbac.ActionCreateUser(), resource: data1, want: false},
		{userID: aliceID, action: domainrbac.ActionViewUser(), resource: data2, want: false},
		{userID: bobID, action: domainrbac.ActionCreateUser(), resource: data2, want: true},
		{userID: bobID, action: domainrbac.ActionViewUser(), resource: data1, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user:%s,%s,%s,want:%v", tt.userID.String(), tt.action.Value(), tt.resource.Value(), tt.want), func(t *testing.T) {
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
	aliceID := randAppUserID(t)

	readerGroup := mustGroup(t, fmt.Sprintf("org:%s,reader", orgID.String()))
	data1 := mustResource(t, fmt.Sprintf("org:%s,data:1", orgID.String()))
	child1 := mustResource(t, fmt.Sprintf("org:%s,child:1", orgID.String()))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, readerGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domainrbac.ActionViewUser(), data1, domainrbac.EffectAllow()))
	require.NoError(t, rbacRepo.AddObjectGroupingPolicy(ctx, orgID, child1, data1))

	tests := []struct {
		resource domainrbac.Resource
		want     bool
	}{
		{resource: data1, want: true},
		{resource: child1, want: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s,want:%v", tt.resource.Value(), tt.want), func(t *testing.T) {
			t.Parallel()
			ok, err := rbacRepo.Enforce(orgID, aliceID, domainrbac.ActionViewUser(), tt.resource)
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
	aliceID := randAppUserID(t)
	readerGroup := mustGroup(t, fmt.Sprintf("org:%s,reader", orgID.String()))

	resources := make([]domainrbac.Resource, 5)
	for i := range 5 {
		resources[i] = mustResource(t, fmt.Sprintf("org:%s,data:%d", orgID.String(), i+1))
	}

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, aliceID, readerGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domainrbac.ActionViewUser(), resources[1], domainrbac.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, readerGroup, domainrbac.ActionViewUser(), resources[3], domainrbac.EffectDeny()))

	// Build hierarchy: data:2 is child of data:1, data:3 of data:2, etc.
	for i := 1; i < 5; i++ {
		require.NoError(t, rbacRepo.AddObjectGroupingPolicy(ctx, orgID, resources[i], resources[i-1]))
	}

	tests := []struct {
		resource domainrbac.Resource
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
			ok, err := rbacRepo.Enforce(orgID, aliceID, domainrbac.ActionViewUser(), tt.resource)
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
	userID := randAppUserID(t)
	adminGroup := mustGroup(t, fmt.Sprintf("org:%s,admin", orgID.String()))

	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, adminGroup))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, adminGroup, domainrbac.ActionCreateUser(), domainrbac.ResourceAny(), domainrbac.EffectAllow()))
	require.NoError(t, rbacRepo.AddPolicy(ctx, orgID, adminGroup, domainrbac.ActionViewUser(), domainrbac.ResourceAny(), domainrbac.EffectAllow()))

	// when
	ok, err := checker.IsAllowed(ctx, orgID, userID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny())

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
	userID := randAppUserID(t)

	// when (no group assigned)
	ok, err := checker.IsAllowed(ctx, orgID, userID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny())

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
	userID := randAppUserID(t)

	adminGroup := mustGroup(t, fmt.Sprintf("org:%s,admin", orgID.String()))
	editorGroup := mustGroup(t, fmt.Sprintf("org:%s,editor", orgID.String()))

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
	userID := randAppUserID(t)

	// when
	groups, err := rbacRepo.GetGroupsForUser(ctx, orgID, userID)

	// then
	require.NoError(t, err)
	assert.Empty(t, groups)
}
