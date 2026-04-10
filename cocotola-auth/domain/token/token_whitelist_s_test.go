package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

func Test_TokenWhitelist_Add_shouldReturnNoEvictions_whenUnderMaxSize(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, existing, 3)
	require.NoError(t, err)

	newEntry := token.WhitelistEntry{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Nil(t, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldReturnOldestID_whenAtMaxSize(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, existing, 3)
	require.NoError(t, err)

	newEntry := token.WhitelistEntry{ID: "token-4", CreatedAt: now.Add(3 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Equal(t, []string{"token-1"}, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldEvictByCreatedAt_whenEntriesAreUnordered(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []token.WhitelistEntry{
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, existing, 2)
	require.NoError(t, err)

	newEntry := token.WhitelistEntry{ID: "token-4", CreatedAt: now.Add(3 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Equal(t, []string{"token-1", "token-2"}, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldEvictExisting_whenMaxSizeIsOne(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, existing, 1)
	require.NoError(t, err)

	newEntry := token.WhitelistEntry{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Equal(t, []string{"token-1"}, evictedIDs)
}

func Test_TokenWhitelist_NewTokenWhitelist_shouldNotLeakMutation_whenOriginalSliceIsModified(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 3)
	require.NoError(t, err)

	// when - mutate original slice
	entries[0] = token.WhitelistEntry{ID: "mutated", CreatedAt: now}
	evictedIDs := whitelist.Add(token.WhitelistEntry{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)})

	// then - whitelist should not be affected by external mutation
	assert.Nil(t, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldReturnNoEvictions_whenAddingToEmptyWhitelist(t *testing.T) {
	t.Parallel()

	// given
	whitelist, err := token.NewWhitelist(fixtureAppUserID, nil, 3)
	require.NoError(t, err)
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	newEntry := token.WhitelistEntry{ID: "token-1", CreatedAt: now}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Nil(t, evictedIDs)
}

func Test_TokenWhitelist_UserID_shouldReturnUserID(t *testing.T) {
	t.Parallel()

	// given
	whitelist, err := token.NewWhitelist(fixtureAppUserID, nil, 3)
	require.NoError(t, err)

	// when
	userID := whitelist.UserID()

	// then
	assert.True(t, fixtureAppUserID.Equal(userID))
}

func Test_TokenWhitelist_Entries_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 3)
	require.NoError(t, err)

	// when
	result := whitelist.Entries()
	result[0] = token.WhitelistEntry{ID: "mutated", CreatedAt: now}

	// then - mutation of returned slice should not affect whitelist
	assert.Equal(t, "token-1", whitelist.Entries()[0].ID)
}

func Test_TokenWhitelist_Entries_shouldReturnAllEntries(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 3)
	require.NoError(t, err)

	// when
	result := whitelist.Entries()

	// then
	assert.Len(t, result, 2)
	assert.Equal(t, "token-1", result[0].ID)
	assert.Equal(t, "token-2", result[1].ID)
}

func Test_TokenWhitelist_ContainsToken_shouldReturnTrue_whenTokenExists(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 3)
	require.NoError(t, err)

	// when
	exists := whitelist.ContainsToken("token-2")

	// then
	assert.True(t, exists)
}

func Test_TokenWhitelist_ContainsToken_shouldReturnFalse_whenTokenDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 3)
	require.NoError(t, err)

	// when
	exists := whitelist.ContainsToken("token-999")

	// then
	assert.False(t, exists)
}

func Test_TokenWhitelist_Remove_shouldRemoveSingleEntry(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 5)
	require.NoError(t, err)

	// when
	whitelist.Remove([]string{"token-2"})

	// then
	result := whitelist.Entries()
	assert.Len(t, result, 2)
	assert.Equal(t, "token-1", result[0].ID)
	assert.Equal(t, "token-3", result[1].ID)
}

func Test_TokenWhitelist_Remove_shouldRemoveMultipleEntries(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 5)
	require.NoError(t, err)

	// when
	whitelist.Remove([]string{"token-1", "token-3"})

	// then
	result := whitelist.Entries()
	assert.Len(t, result, 1)
	assert.Equal(t, "token-2", result[0].ID)
}

func Test_TokenWhitelist_Remove_shouldDoNothing_whenIDDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 5)
	require.NoError(t, err)

	// when
	whitelist.Remove([]string{"non-existent"})

	// then
	assert.Len(t, whitelist.Entries(), 1)
}

func Test_TokenWhitelist_Remove_shouldDoNothing_whenEmptySlice(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []token.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist, err := token.NewWhitelist(fixtureAppUserID, entries, 5)
	require.NoError(t, err)

	// when
	whitelist.Remove([]string{})

	// then
	assert.Len(t, whitelist.Entries(), 1)
}

func Test_NewWhitelist_shouldReturnError_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	// when
	_, err := token.NewWhitelist(domain.AppUserID{}, nil, 3)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user id")
}

func Test_NewWhitelist_shouldReturnError_whenMaxSizeIsZero(t *testing.T) {
	t.Parallel()

	// when
	_, err := token.NewWhitelist(fixtureAppUserID, nil, 0)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max size must be positive")
}

func Test_NewWhitelist_shouldReturnError_whenMaxSizeIsNegative(t *testing.T) {
	t.Parallel()

	// when
	_, err := token.NewWhitelist(fixtureAppUserID, nil, -1)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max size must be positive")
}
