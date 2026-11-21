package authz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/service/authz"
)

var (
	fixtureOrgID  = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000573")
	fixtureUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000574")
)

func Test_NewAddPolicyInput_shouldReturnInputWithGivenArguments(t *testing.T) {
	t.Parallel()

	// given
	resource, err := domainrbac.ResourceWorkbook("wb-573")
	require.NoError(t, err)

	// when
	got := authz.NewAddPolicyInput(fixtureOrgID, fixtureUserID, domainrbac.ActionCreateQuestion(), resource, domainrbac.EffectDeny())

	// then
	assert.Equal(t, &authz.AddPolicyInput{
		OrganizationID: fixtureOrgID,
		AppUserID:      fixtureUserID,
		Action:         domainrbac.ActionCreateQuestion(),
		Resource:       resource,
		Effect:         domainrbac.EffectDeny(),
	}, got)
}
