//go:build medium

package initialize_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

func randUserID(t *testing.T) domain.AppUserID {
	t.Helper()
	id, err := domain.NewAppUserIDV7()
	require.NoError(t, err)
	return id
}

func randSpaceID(t *testing.T) domain.SpaceID {
	t.Helper()
	id, err := domain.NewSpaceIDV7()
	require.NoError(t, err)
	return id
}

func newRBACRepository(t *testing.T) *gateway.RBACRepository {
	t.Helper()
	repo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	return repo
}

func adminActions() []domainrbac.Action {
	return []domainrbac.Action{
		domainrbac.ActionCreateUser(),
		domainrbac.ActionListUsers(),
		domainrbac.ActionViewUser(),
		domainrbac.ActionDisableUser(),
		domainrbac.ActionChangePassword(),
		domainrbac.ActionCreateGroup(),
		domainrbac.ActionListGroups(),
		domainrbac.ActionViewGroup(),
		domainrbac.ActionDisableGroup(),
		domainrbac.ActionAddUserToGroup(),
		domainrbac.ActionRemoveUserFromGroup(),
		domainrbac.ActionCreateSpace(),
		domainrbac.ActionListSpaces(),
		domainrbac.ActionViewSpace(),
	}
}

func Test_setupRBACPolicies_shouldAllowAdminActions_whenAdminGroupIsAssigned(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	orgID := randOrgID(t)
	userID := randUserID(t)
	repo := newRBACRepository(t)
	require.NoError(t, initialize.SetupRBACPolicies(ctx, repo, orgID, slog.Default()))
	require.NoError(t, initialize.AssignGroup(ctx, repo, orgID, userID, domainrbac.GroupAdmin(), slog.Default()))
	enforcer := newRBACRepository(t)

	for _, action := range adminActions() {
		t.Run(action.Value(), func(t *testing.T) {
			t.Parallel()

			// when
			allowed, err := enforcer.Enforce(orgID, userID, action, domainrbac.ResourceAny())

			// then
			require.NoError(t, err)
			assert.True(t, allowed)
		})
	}
}

func Test_setupRBACPolicies_shouldNotAllowCreateOrganization_whenAdminGroupIsAssigned(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	orgID := randOrgID(t)
	userID := randUserID(t)
	repo := newRBACRepository(t)
	require.NoError(t, initialize.SetupRBACPolicies(ctx, repo, orgID, slog.Default()))
	require.NoError(t, initialize.AssignGroup(ctx, repo, orgID, userID, domainrbac.GroupAdmin(), slog.Default()))
	enforcer := newRBACRepository(t)

	// when
	allowed, err := enforcer.Enforce(orgID, userID, domainrbac.ActionCreateOrganization(), domainrbac.ResourceAny())

	// then
	require.NoError(t, err)
	assert.False(t, allowed)
}

func Test_setupSystemOwnerRBACPolicies_shouldAllowSystemOwnerActions_whenSystemOwnerGroupIsAssigned(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	orgID := randOrgID(t)
	userID := randUserID(t)
	repo := newRBACRepository(t)
	require.NoError(t, initialize.SetupSystemOwnerRBACPolicies(ctx, repo, orgID, slog.Default()))
	require.NoError(t, initialize.AssignGroup(ctx, repo, orgID, userID, domainrbac.GroupSystemOwner(), slog.Default()))
	enforcer := newRBACRepository(t)

	for _, action := range append(adminActions(), domainrbac.ActionCreateOrganization()) {
		t.Run(action.Value(), func(t *testing.T) {
			t.Parallel()

			// when
			allowed, err := enforcer.Enforce(orgID, userID, action, domainrbac.ResourceAny())

			// then
			require.NoError(t, err)
			assert.True(t, allowed)
		})
	}
}

func Test_grantPublicSpacePoliciesToSystemOwner_shouldAllowWorkbookActionsOnPublicSpace_whenSystemOwnerGroupIsAssigned(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	orgID := randOrgID(t)
	userID := randUserID(t)
	spaceID := randSpaceID(t)
	repo := newRBACRepository(t)
	require.NoError(t, initialize.GrantPublicSpacePoliciesToSystemOwner(ctx, repo, orgID, spaceID, slog.Default()))
	require.NoError(t, initialize.AssignGroup(ctx, repo, orgID, userID, domainrbac.GroupSystemOwner(), slog.Default()))
	enforcer := newRBACRepository(t)

	actions := []domainrbac.Action{
		domainrbac.ActionCreateWorkbook(),
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionUpdateWorkbook(),
		domainrbac.ActionDeleteWorkbook(),
		domainrbac.ActionCreateQuestion(),
		domainrbac.ActionUpdateQuestion(),
		domainrbac.ActionDeleteQuestion(),
	}
	for _, action := range actions {
		t.Run(action.Value(), func(t *testing.T) {
			t.Parallel()

			// when
			allowed, err := enforcer.Enforce(orgID, userID, action, domainrbac.ResourceSpace(spaceID))

			// then
			require.NoError(t, err)
			assert.True(t, allowed)
		})
	}
}

func Test_assignGroup_shouldStoreGroupName_whenAssigned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		group domainrbac.Group
		want  string
	}{
		{name: "admin", group: domainrbac.GroupAdmin(), want: "admin"},
		{name: "system owner", group: domainrbac.GroupSystemOwner(), want: "system_owner"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			ctx := context.Background()
			orgID := randOrgID(t)
			userID := randUserID(t)
			repo := newRBACRepository(t)

			// when
			require.NoError(t, initialize.AssignGroup(ctx, repo, orgID, userID, tt.group, slog.Default()))

			// then
			groups, err := newRBACRepository(t).GetGroupsForUser(ctx, orgID, userID)
			require.NoError(t, err)
			assert.Equal(t, []string{tt.want}, groups)
		})
	}
}

func Test_grantPublicSpacePoliciesToSystemOwner_shouldSucceed_whenRunAgainByAnotherProcess(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	orgID := randOrgID(t)
	spaceID := randSpaceID(t)
	require.NoError(t, initialize.GrantPublicSpacePoliciesToSystemOwner(ctx, newRBACRepository(t), orgID, spaceID, slog.Default()))

	// when
	err := initialize.GrantPublicSpacePoliciesToSystemOwner(ctx, newRBACRepository(t), orgID, spaceID, slog.Default())

	// then
	require.NoError(t, err)
}

func Test_grantPublicSpacePoliciesToSystemOwner_shouldNotAllowWorkbookActionsOnOtherSpace_whenSystemOwnerGroupIsAssigned(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	orgID := randOrgID(t)
	userID := randUserID(t)
	repo := newRBACRepository(t)
	require.NoError(t, initialize.GrantPublicSpacePoliciesToSystemOwner(ctx, repo, orgID, randSpaceID(t), slog.Default()))
	require.NoError(t, initialize.AssignGroup(ctx, repo, orgID, userID, domainrbac.GroupSystemOwner(), slog.Default()))
	enforcer := newRBACRepository(t)

	// when
	allowed, err := enforcer.Enforce(orgID, userID, domainrbac.ActionCreateWorkbook(), domainrbac.ResourceSpace(randSpaceID(t)))

	// then
	require.NoError(t, err)
	assert.False(t, allowed)
}
