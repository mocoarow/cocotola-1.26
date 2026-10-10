//go:build medium

package initialize_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

// orgNameSuffixLength keeps the public space key name ("public@@" + name)
// within its 50-character limit while staying unique per test.
const orgNameSuffixLength = 12

// newOrganizationWithPublicSpace saves an organization unique to the test and
// provisions its public space.
func newOrganizationWithPublicSpace(ctx context.Context, t *testing.T) (domain.OrganizationID, *domainspace.Space) {
	t.Helper()
	orgID := randOrgID(t)
	id := orgID.String()
	name := "finder-" + id[len(id)-orgNameSuffixLength:]
	org, err := domain.NewOrganization(orgID, name, 10, 10)
	require.NoError(t, err)
	require.NoError(t, gateway.NewOrganizationRepository(testDB).Save(ctx, org))

	space, err := domainspace.Provision(ctx, gateway.NewSpaceRepository(testDB), orgID, domain.SystemAppUserID(),
		domainspace.PublicSpaceKeyName(name), "Public", domainspace.TypePublic())
	require.NoError(t, err)
	return orgID, space
}

func Test_PublicSpaceFinder_FindPublicSpaceID_shouldReturnSpaceID_whenPublicSpaceExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID, space := newOrganizationWithPublicSpace(ctx, t)
	finder := initialize.NewPublicSpaceFinder(testDB)

	// when
	got, err := finder.FindPublicSpaceID(ctx, orgID)

	// then
	require.NoError(t, err)
	assert.Equal(t, space.ID(), got)
}

func Test_PublicSpaceFinder_FindPublicSpaceID_shouldReturnErrPublicSpaceNotFound_whenSpaceMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: an organization ID that has no spaces
	finder := initialize.NewPublicSpaceFinder(testDB)

	// when
	_, err := finder.FindPublicSpaceID(ctx, randOrgID(t))

	// then
	require.ErrorIs(t, err, initialize.ErrPublicSpaceNotFound)
}

func Test_PublicSpaceFinder_FindPublicSpaceID_shouldReturnErrPublicSpaceNotFound_whenPublicSpaceDeleted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID, space := newOrganizationWithPublicSpace(ctx, t)
	space.Delete()
	require.NoError(t, gateway.NewSpaceRepository(testDB).Save(ctx, space))
	finder := initialize.NewPublicSpaceFinder(testDB)

	// when
	_, err := finder.FindPublicSpaceID(ctx, orgID)

	// then
	require.ErrorIs(t, err, initialize.ErrPublicSpaceNotFound)
}
