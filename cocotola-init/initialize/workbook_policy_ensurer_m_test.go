//go:build medium

package initialize_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

func randOrgID(t *testing.T) domain.OrganizationID {
	t.Helper()
	id, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	return id
}

func mustAction(t *testing.T, value string) domainrbac.Action {
	t.Helper()
	action, err := domainrbac.NewAction(value)
	require.NoError(t, err)
	return action
}

func mustResourceWorkbook(t *testing.T, workbookID string) domainrbac.Resource {
	t.Helper()
	resource, err := domainrbac.ResourceWorkbook(workbookID)
	require.NoError(t, err)
	return resource
}

func Test_WorkbookPolicyEnsurer_EnsureSystemOwnerWorkbookPolicies_shouldGrantAllWorkbookActions_whenCalled(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	ensurer, err := initialize.NewWorkbookPolicyEnsurer(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	workbookID := fmt.Sprintf("wb-%s", orgID.String())

	// when
	err = ensurer.EnsureSystemOwnerWorkbookPolicies(ctx, orgID.String(), workbookID)

	// then: the system owner is allowed every workbook-scoped action the seeder needs.
	// A fresh repository is built after the grant so its enforcer loads the rows
	// the ensurer just persisted.
	require.NoError(t, err)
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	resource := mustResourceWorkbook(t, workbookID)
	systemOwnerID := domain.SystemAppUserID()

	tests := []struct {
		action domainrbac.Action
	}{
		{action: domainrbac.ActionViewWorkbook()},
		{action: domainrbac.ActionUpdateWorkbook()},
		{action: domainrbac.ActionDeleteWorkbook()},
		{action: mustAction(t, "study_workbook")},
		{action: domainrbac.ActionCreateQuestion()},
		{action: domainrbac.ActionUpdateQuestion()},
		{action: domainrbac.ActionDeleteQuestion()},
	}
	for _, tt := range tests {
		t.Run(tt.action.Value(), func(t *testing.T) {
			t.Parallel()
			allowed, err := rbacRepo.Enforce(orgID, systemOwnerID, tt.action, resource)
			require.NoError(t, err)
			assert.True(t, allowed)
		})
	}
}

func Test_WorkbookPolicyEnsurer_EnsureSystemOwnerWorkbookPolicies_shouldNotGrantUnrelatedAction_whenCalled(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	ensurer, err := initialize.NewWorkbookPolicyEnsurer(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	workbookID := fmt.Sprintf("wb-%s", orgID.String())

	// when
	err = ensurer.EnsureSystemOwnerWorkbookPolicies(ctx, orgID.String(), workbookID)

	// then: create_workbook is not part of the granted set, so it stays denied
	require.NoError(t, err)
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	allowed, err := rbacRepo.Enforce(orgID, domain.SystemAppUserID(), domainrbac.ActionCreateWorkbook(), mustResourceWorkbook(t, workbookID))
	require.NoError(t, err)
	assert.False(t, allowed)
}

func Test_WorkbookPolicyEnsurer_EnsureSystemOwnerWorkbookPolicies_shouldBeIdempotent_whenCalledTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	ensurer, err := initialize.NewWorkbookPolicyEnsurer(testDB)
	require.NoError(t, err)

	orgID := randOrgID(t)
	workbookID := fmt.Sprintf("wb-%s", orgID.String())

	// when: the same workbook is ensured twice, mirroring a repeated init run
	require.NoError(t, ensurer.EnsureSystemOwnerWorkbookPolicies(ctx, orgID.String(), workbookID))
	err = ensurer.EnsureSystemOwnerWorkbookPolicies(ctx, orgID.String(), workbookID)

	// then: the second run succeeds and the grant is still effective (no duplicate rows)
	require.NoError(t, err)
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	allowed, err := rbacRepo.Enforce(orgID, domain.SystemAppUserID(), domainrbac.ActionCreateQuestion(), mustResourceWorkbook(t, workbookID))
	require.NoError(t, err)
	assert.True(t, allowed)
}
