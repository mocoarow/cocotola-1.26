//go:build medium

package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

const whitelistMaxSize = 3

var whitelistBaseTime = time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)

type whitelistRepository interface {
	FindByUserID(ctx context.Context, userID domain.AppUserID) (*domaintoken.Whitelist, error)
	Save(ctx context.Context, whitelist *domaintoken.Whitelist) error
}

func saveSessionTokenIDs(ctx context.Context, t *testing.T, db *gorm.DB, userID domain.AppUserID, count int) []string {
	t.Helper()
	repo := gateway.NewSessionTokenRepository(db)
	ids := make([]string, count)
	for i := range count {
		token := newTestSessionToken(t, userID)
		require.NoError(t, repo.Save(ctx, token))
		ids[i] = token.ID()
	}
	return ids
}

func saveRefreshTokenID(ctx context.Context, t *testing.T, db *gorm.DB, userID domain.AppUserID) string {
	t.Helper()
	token := newTestRefreshToken(t, userID)
	require.NoError(t, gateway.NewRefreshTokenRepository(db).Save(ctx, token))
	return token.ID()
}

func saveAccessTokenID(ctx context.Context, t *testing.T, db *gorm.DB, userID domain.AppUserID) string {
	t.Helper()
	token := domaintoken.ReconstructAccessToken(uuid.NewString(), saveRefreshTokenID(ctx, t, db, userID), userID, "access-user@example.com", fixtureOrganizationName, opaqueTokenCreatedAt, opaqueTokenExpiresAt, nil)
	require.NoError(t, gateway.NewAccessTokenRepository(db).Save(ctx, token))
	return token.ID()
}

func whitelistEntryAt(id string, minutes int) domaintoken.WhitelistEntry {
	return domaintoken.WhitelistEntry{ID: id, CreatedAt: whitelistBaseTime.Add(time.Duration(minutes) * time.Minute)}
}

func whitelistEntryIDs(whitelist *domaintoken.Whitelist) []string {
	entries := whitelist.Entries()
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}

func findWhitelist(ctx context.Context, t *testing.T, repo whitelistRepository, userID domain.AppUserID) *domaintoken.Whitelist {
	t.Helper()
	whitelist, err := repo.FindByUserID(ctx, userID)
	require.NoError(t, err)
	return whitelist
}

func saveWhitelistWith(ctx context.Context, t *testing.T, repo whitelistRepository, userID domain.AppUserID, ids ...string) {
	t.Helper()
	whitelist := findWhitelist(ctx, t, repo, userID)
	for i, id := range ids {
		require.NoError(t, whitelist.Add(whitelistEntryAt(id, i), whitelistMaxSize))
	}
	require.NoError(t, repo.Save(ctx, whitelist))
}

func Test_TokenWhitelistRepositories_Save_shouldPersistEntries_whenFoundAgain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		newRepo func(db *gorm.DB) whitelistRepository
		saveID  func(ctx context.Context, t *testing.T, db *gorm.DB, userID domain.AppUserID) string
	}{
		{
			name:    "session",
			newRepo: func(db *gorm.DB) whitelistRepository { return gateway.NewSessionTokenWhitelistRepository(db) },
			saveID: func(ctx context.Context, t *testing.T, db *gorm.DB, userID domain.AppUserID) string {
				t.Helper()
				return saveSessionTokenIDs(ctx, t, db, userID, 1)[0]
			},
		},
		{
			name:    "refresh",
			newRepo: func(db *gorm.DB) whitelistRepository { return gateway.NewRefreshTokenWhitelistRepository(db) },
			saveID:  saveRefreshTokenID,
		},
		{
			name:    "access",
			newRepo: func(db *gorm.DB) whitelistRepository { return gateway.NewAccessTokenWhitelistRepository(db) },
			saveID:  saveAccessTokenID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// given
			ctx := context.Background()
			tx := testDB.Begin()
			defer tx.Rollback()
			userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-save-"+tt.name)
			repo := tt.newRepo(tx)
			id := tt.saveID(ctx, t, tx, userID)

			// when
			saveWhitelistWith(ctx, t, repo, userID, id)

			// then
			assert.Equal(t, []string{id}, whitelistEntryIDs(findWhitelist(ctx, t, repo, userID)))
		})
	}
}

func Test_SessionTokenWhitelistRepository_FindByUserID_shouldReturnVersionZero_whenNeverSaved(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-unsaved")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)

	// when
	whitelist := findWhitelist(ctx, t, repo, userID)

	// then
	assert.Equal(t, 0, whitelist.Version())
}

func givenLegacySessionWhitelist(ctx context.Context, t *testing.T, tx *gorm.DB, name string) (domain.AppUserID, []string) {
	t.Helper()
	userID := setupOpaqueTokenUser(ctx, t, tx, name)
	ids := saveSessionTokenIDs(ctx, t, tx, userID, 2)
	for i, id := range ids {
		require.NoError(t, tx.Exec("INSERT INTO session_token_whitelist (user_id, token_id, created_at) VALUES (?, ?, ?)",
			userID.String(), id, whitelistBaseTime.Add(time.Duration(i)*time.Minute)).Error)
	}
	return userID, ids
}

func Test_SessionTokenWhitelistRepository_FindByUserID_shouldReturnLegacyEntries_whenRootRowIsMissing(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID, ids := givenLegacySessionWhitelist(ctx, t, tx, "whitelist-legacy-entries")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)

	// when
	whitelist := findWhitelist(ctx, t, repo, userID)

	// then
	assert.ElementsMatch(t, ids, whitelistEntryIDs(whitelist))
}

func Test_SessionTokenWhitelistRepository_FindByUserID_shouldReturnVersionZero_whenRootRowIsMissing(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID, _ := givenLegacySessionWhitelist(ctx, t, tx, "whitelist-legacy-version")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)

	// when
	whitelist := findWhitelist(ctx, t, repo, userID)

	// then
	assert.Equal(t, 0, whitelist.Version())
}

func Test_SessionTokenWhitelistRepository_Save_shouldCreateRootRow_whenRootRowIsMissing(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID, ids := givenLegacySessionWhitelist(ctx, t, tx, "whitelist-legacy-save")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)
	whitelist := findWhitelist(ctx, t, repo, userID)
	whitelist.Remove(ids[:1])

	// when
	err := repo.Save(ctx, whitelist)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, findWhitelist(ctx, t, repo, userID).Version())
}

func Test_SessionTokenWhitelistRepository_FindByUserID_shouldReturnSavedVersion_whenSavedTwice(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-version")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)
	ids := saveSessionTokenIDs(ctx, t, tx, userID, 2)
	saveWhitelistWith(ctx, t, repo, userID, ids[0])
	saveWhitelistWith(ctx, t, repo, userID, ids[1])

	// when
	whitelist := findWhitelist(ctx, t, repo, userID)

	// then
	assert.Equal(t, 2, whitelist.Version())
}

func Test_SessionTokenWhitelistRepository_Save_shouldReplaceEntries_whenEntryIsRemoved(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-replace")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)
	ids := saveSessionTokenIDs(ctx, t, tx, userID, 2)
	saveWhitelistWith(ctx, t, repo, userID, ids...)
	whitelist := findWhitelist(ctx, t, repo, userID)
	whitelist.Remove(ids[:1])

	// when
	err := repo.Save(ctx, whitelist)

	// then
	require.NoError(t, err)
	assert.Equal(t, ids[1:], whitelistEntryIDs(findWhitelist(ctx, t, repo, userID)))
}

func Test_SessionTokenWhitelistRepository_Save_shouldReturnErrConcurrentModification_whenVersionIsStale(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-stale")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)
	ids := saveSessionTokenIDs(ctx, t, tx, userID, 3)
	saveWhitelistWith(ctx, t, repo, userID, ids[0])
	first := findWhitelist(ctx, t, repo, userID)
	second := findWhitelist(ctx, t, repo, userID)
	require.NoError(t, first.Add(whitelistEntryAt(ids[1], 1), whitelistMaxSize))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(whitelistEntryAt(ids[2], 2), whitelistMaxSize))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_SessionTokenWhitelistRepository_Save_shouldReturnErrConcurrentModification_whenRootIsCreatedConcurrently(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-first-save")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)
	ids := saveSessionTokenIDs(ctx, t, tx, userID, 2)
	first := findWhitelist(ctx, t, repo, userID)
	second := findWhitelist(ctx, t, repo, userID)
	require.NoError(t, first.Add(whitelistEntryAt(ids[0], 0), whitelistMaxSize))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(whitelistEntryAt(ids[1], 1), whitelistMaxSize))

	// when
	err := repo.Save(ctx, second)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_SessionTokenWhitelistRepository_Save_shouldKeepEntriesOfWinner_whenLoserConflicts(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	userID := setupOpaqueTokenUser(ctx, t, tx, "whitelist-winner")
	repo := gateway.NewSessionTokenWhitelistRepository(tx)
	ids := saveSessionTokenIDs(ctx, t, tx, userID, 3)
	saveWhitelistWith(ctx, t, repo, userID, ids[0])
	first := findWhitelist(ctx, t, repo, userID)
	second := findWhitelist(ctx, t, repo, userID)
	require.NoError(t, first.Add(whitelistEntryAt(ids[1], 1), whitelistMaxSize))
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, second.Add(whitelistEntryAt(ids[2], 2), whitelistMaxSize))

	// when
	_ = repo.Save(ctx, second)

	// then
	assert.ElementsMatch(t, ids[:2], whitelistEntryIDs(findWhitelist(ctx, t, repo, userID)))
}
