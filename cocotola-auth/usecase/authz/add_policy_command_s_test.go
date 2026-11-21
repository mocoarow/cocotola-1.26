package authz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	authzservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/authz"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/authz"
)

var (
	fixtureOrgID  = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000573")
	fixtureUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000574")
)

func newFixtureInput(t *testing.T) *authzservice.AddPolicyInput {
	t.Helper()
	resource, err := domainrbac.ResourceWorkbook("wb-573")
	require.NoError(t, err)
	return &authzservice.AddPolicyInput{
		OrganizationID: fixtureOrgID,
		AppUserID:      fixtureUserID,
		Action:         domainrbac.ActionCreateQuestion(),
		Resource:       resource,
		Effect:         domainrbac.EffectDeny(),
	}
}

func Test_AddPolicyCommand_AddPolicy_shouldAddPolicyWithInputValues_whenInputIsValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	input := newFixtureInput(t)
	policyAdder := newMockuserPolicyAdder(t)
	policyAdder.On("AddPolicyForUser", mock.Anything, input.OrganizationID, input.AppUserID, input.Action, input.Resource, input.Effect).Return(nil)
	cmd := authz.NewAddPolicyCommand(policyAdder)

	// when
	err := cmd.AddPolicy(ctx, input)

	// then
	require.NoError(t, err)
}

func Test_AddPolicyCommand_AddPolicy_shouldReturnError_whenPolicyAdderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	errAdd := errors.New("db error")
	policyAdder := newMockuserPolicyAdder(t)
	policyAdder.On("AddPolicyForUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errAdd)
	cmd := authz.NewAddPolicyCommand(policyAdder)

	// when
	err := cmd.AddPolicy(ctx, newFixtureInput(t))

	// then
	require.ErrorIs(t, err, errAdd)
}
