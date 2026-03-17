package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_TokenWhitelist_Add_shouldReturnNoEvictions_whenUnderMaxSize(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist := domain.NewTokenWhitelist(1, existing, 3)

	newEntry := domain.WhitelistEntry{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Nil(t, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldReturnOldestID_whenAtMaxSize(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, existing, 3)

	newEntry := domain.WhitelistEntry{ID: "token-4", CreatedAt: now.Add(3 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Equal(t, []string{"token-1"}, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldEvictByCreatedAt_whenEntriesAreUnordered(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []domain.WhitelistEntry{
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, existing, 2)

	newEntry := domain.WhitelistEntry{ID: "token-4", CreatedAt: now.Add(3 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Equal(t, []string{"token-1", "token-2"}, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldEvictExisting_whenMaxSizeIsOne(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist := domain.NewTokenWhitelist(1, existing, 1)

	newEntry := domain.WhitelistEntry{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Equal(t, []string{"token-1"}, evictedIDs)
}

func Test_TokenWhitelist_NewTokenWhitelist_shouldNotLeakMutation_whenOriginalSliceIsModified(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 3)

	// when - mutate original slice
	entries[0] = domain.WhitelistEntry{ID: "mutated", CreatedAt: now}
	evictedIDs := whitelist.Add(domain.WhitelistEntry{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)})

	// then - whitelist should not be affected by external mutation
	assert.Nil(t, evictedIDs)
}

func Test_TokenWhitelist_Add_shouldReturnNoEvictions_whenAddingToEmptyWhitelist(t *testing.T) {
	t.Parallel()

	// given
	whitelist := domain.NewTokenWhitelist(1, nil, 3)
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	newEntry := domain.WhitelistEntry{ID: "token-1", CreatedAt: now}

	// when
	evictedIDs := whitelist.Add(newEntry)

	// then
	assert.Nil(t, evictedIDs)
}

func Test_TokenWhitelist_UserID_shouldReturnUserID(t *testing.T) {
	t.Parallel()

	// given
	whitelist := domain.NewTokenWhitelist(42, nil, 3)

	// when
	userID := whitelist.UserID()

	// then
	assert.Equal(t, 42, userID)
}

func Test_TokenWhitelist_Entries_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 3)

	// when
	result := whitelist.Entries()
	result[0] = domain.WhitelistEntry{ID: "mutated", CreatedAt: now}

	// then - mutation of returned slice should not affect whitelist
	assert.Equal(t, "token-1", whitelist.Entries()[0].ID)
}

func Test_TokenWhitelist_Entries_shouldReturnAllEntries(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 3)

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
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 3)

	// when
	exists := whitelist.ContainsToken("token-2")

	// then
	assert.True(t, exists)
}

func Test_TokenWhitelist_ContainsToken_shouldReturnFalse_whenTokenDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 3)

	// when
	exists := whitelist.ContainsToken("token-999")

	// then
	assert.False(t, exists)
}

func Test_TokenWhitelist_Remove_shouldRemoveSingleEntry(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 5)

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
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
		{ID: "token-2", CreatedAt: now.Add(1 * time.Minute)},
		{ID: "token-3", CreatedAt: now.Add(2 * time.Minute)},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 5)

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
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 5)

	// when
	whitelist.Remove([]string{"non-existent"})

	// then
	assert.Len(t, whitelist.Entries(), 1)
}

func Test_TokenWhitelist_Remove_shouldDoNothing_whenEmptySlice(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.WhitelistEntry{
		{ID: "token-1", CreatedAt: now},
	}
	whitelist := domain.NewTokenWhitelist(1, entries, 5)

	// when
	whitelist.Remove([]string{})

	// then
	assert.Len(t, whitelist.Entries(), 1)
}
