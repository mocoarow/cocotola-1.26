package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

var (
	fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")
	fixtureOrgID     = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")
)

const fixtureWhitelistSize = 2

func fullTestWhitelist(t *testing.T, now time.Time) *domaintoken.Whitelist {
	t.Helper()
	return newTestWhitelist(t,
		domaintoken.WhitelistEntry{ID: "oldest", CreatedAt: now.Add(-2 * time.Hour)},
		domaintoken.WhitelistEntry{ID: "newer", CreatedAt: now.Add(-time.Hour)},
	)
}

func whitelistEntryIDs(whitelist *domaintoken.Whitelist) []string {
	entries := whitelist.Entries()
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}

func newTestWhitelist(t *testing.T, entries ...domaintoken.WhitelistEntry) *domaintoken.Whitelist {
	t.Helper()
	whitelist, err := domaintoken.NewWhitelist(fixtureAppUserID, entries)
	require.NoError(t, err)
	return whitelist
}
