package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

var whitelistFixtureNow = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func whitelistEntry(id string, minutes int) token.WhitelistEntry {
	return token.WhitelistEntry{ID: id, CreatedAt: whitelistFixtureNow.Add(time.Duration(minutes) * time.Minute)}
}

func entryIDs(entries []token.WhitelistEntry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}

func newWhitelist(t *testing.T, entries ...token.WhitelistEntry) *token.Whitelist {
	t.Helper()
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries)
	require.NoError(t, err)
	return whitelist
}

func Test_TokenWhitelist_Add_shouldKeepAllEntries_whenUnderMaxSize(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0))

	// when
	err := whitelist.Add(whitelistEntry("token-2", 1), 3)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"token-1", "token-2"}, entryIDs(whitelist.Entries()))
}

func Test_TokenWhitelist_Add_shouldEvictOldestByCreatedAt_whenMaxSizeExceeded(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-3", 2), whitelistEntry("token-1", 0), whitelistEntry("token-2", 1))

	// when
	err := whitelist.Add(whitelistEntry("token-4", 3), 2)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"token-3", "token-4"}, entryIDs(whitelist.Entries()))
}

func Test_TokenWhitelist_Add_shouldReturnErrInvalidArgument_whenMaxSizeIsNotPositive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		maxSize int
	}{
		{name: "zero", maxSize: 0},
		{name: "negative", maxSize: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			whitelist := newWhitelist(t)

			// when
			err := whitelist.Add(whitelistEntry("token-1", 0), tt.maxSize)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_TokenWhitelist_Add_shouldReturnErrDuplicateEntry_whenTokenIsAlreadyListed(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0))

	// when
	err := whitelist.Add(whitelistEntry("token-1", 1), 3)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_TokenWhitelist_Add_shouldKeepEntries_whenTokenIsAlreadyListed(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0))

	// when
	_ = whitelist.Add(whitelistEntry("token-1", 1), 3)

	// then
	assert.Equal(t, []token.WhitelistEntry{whitelistEntry("token-1", 0)}, whitelist.Entries())
}

func Test_TokenWhitelist_Remove_shouldRemoveGivenEntries_whenTheyExist(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0), whitelistEntry("token-2", 1), whitelistEntry("token-3", 2))

	// when
	whitelist.Remove([]string{"token-1", "token-3"})

	// then
	assert.Equal(t, []string{"token-2"}, entryIDs(whitelist.Entries()))
}

func Test_TokenWhitelist_Remove_shouldKeepEntries_whenIDDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0))

	// when
	whitelist.Remove([]string{"non-existent"})

	// then
	assert.Equal(t, []string{"token-1"}, entryIDs(whitelist.Entries()))
}

func Test_TokenWhitelist_ContainsToken_shouldReturnTrue_whenTokenExists(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0), whitelistEntry("token-2", 1))

	// when
	exists := whitelist.ContainsToken("token-2")

	// then
	assert.True(t, exists)
}

func Test_TokenWhitelist_ContainsToken_shouldReturnFalse_whenTokenDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0))

	// when
	exists := whitelist.ContainsToken("token-999")

	// then
	assert.False(t, exists)
}

func Test_TokenWhitelist_Entries_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t, whitelistEntry("token-1", 0))

	// when
	whitelist.Entries()[0] = whitelistEntry("mutated", 0)

	// then
	assert.Equal(t, []string{"token-1"}, entryIDs(whitelist.Entries()))
}

func Test_TokenWhitelist_UserID_shouldReturnUserID(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t)

	// when
	userID := whitelist.UserID()

	// then
	assert.True(t, fixtureAppUserID.Equal(userID))
}

func Test_TokenWhitelist_Version_shouldReturnZero_whenNew(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t)

	// when
	version := whitelist.Version()

	// then
	assert.Equal(t, 0, version)
}

func Test_TokenWhitelist_SetVersion_shouldUpdateVersion(t *testing.T) {
	t.Parallel()

	// given
	whitelist := newWhitelist(t)

	// when
	whitelist.SetVersion(7)

	// then
	assert.Equal(t, 7, whitelist.Version())
}

func Test_NewWhitelist_shouldNotShareEntries_whenOriginalSliceIsModified(t *testing.T) {
	t.Parallel()

	// given
	entries := []token.WhitelistEntry{whitelistEntry("token-1", 0)}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries)
	require.NoError(t, err)

	// when
	entries[0] = whitelistEntry("mutated", 0)

	// then
	assert.Equal(t, []string{"token-1"}, entryIDs(whitelist.Entries()))
}

func Test_NewWhitelist_shouldReturnErrInvalidArgument_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	zeroUserID := domain.AppUserID{}

	// when
	_, err := token.NewWhitelist(zeroUserID, nil)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
