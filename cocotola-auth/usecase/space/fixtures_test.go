package space_test

import (
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

const (
	fixtureOrgName        = "acme"
	fixturePublicSpaceKey = "public@@acme"
	fixtureSpaceName      = "Public Space"
)

var (
	fixtureSpaceID = domain.MustParseSpaceID("11111111-1111-7111-8111-111111111111")
	fixtureOrgID   = domain.MustParseOrganizationID("22222222-2222-7222-8222-222222222222")
	fixtureOwnerID = domain.MustParseAppUserID("33333333-3333-7333-8333-333333333333")
)

func fixtureOrganization(t *testing.T) *domain.Organization {
	t.Helper()
	return domain.ReconstructOrganization(fixtureOrgID, fixtureOrgName, 100, 100)
}
