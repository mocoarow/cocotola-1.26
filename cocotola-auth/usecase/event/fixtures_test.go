package event_test

import (
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var (
	fixtureOrgID     = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")
	fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")
	fixtureUser1     = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	fixtureUser2     = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000022")
)
