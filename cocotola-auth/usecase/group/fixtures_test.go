package group_test

import (
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

const (
	fixtureOrgName   = "acme"
	fixtureGroupName = "engineers"
)

var (
	fixtureOrgID      = domain.MustParseOrganizationID("22222222-2222-7222-8222-222222222222")
	fixtureOperatorID = domain.MustParseAppUserID("33333333-3333-7333-8333-333333333333")
)

func fixtureOrganization(t *testing.T) *domain.Organization {
	t.Helper()
	return domain.ReconstructOrganization(fixtureOrgID, fixtureOrgName, 100, 100)
}
