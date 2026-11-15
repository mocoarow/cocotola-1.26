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
)

const fixtureOrganizationName = "fixture-org"

var (
	opaqueTokenCreatedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	opaqueTokenExpiresAt = time.Date(2026, 1, 9, 3, 4, 5, 0, time.UTC)
	opaqueTokenRevokedAt = time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)
)

func setupOpaqueTokenUser(ctx context.Context, t *testing.T, tx *gorm.DB, orgName string) domain.AppUserID {
	t.Helper()
	orgID := setupOrganization(ctx, t, tx, orgName)
	return setupUsers(ctx, t, tx, orgID, orgName, 1)[0]
}

func newTestTokenHash(t *testing.T) domain.TokenHash {
	t.Helper()
	_, hash, err := domaintoken.GenerateOpaqueToken()
	require.NoError(t, err)
	return hash
}

func newTestRefreshToken(t *testing.T, userID domain.AppUserID) *domaintoken.RefreshToken {
	t.Helper()
	return domaintoken.ReconstructRefreshToken(uuid.NewString(), userID, "refresh-user@example.com", fixtureOrganizationName, newTestTokenHash(t), opaqueTokenCreatedAt, opaqueTokenExpiresAt, nil)
}

func newTestSessionToken(t *testing.T, userID domain.AppUserID) *domaintoken.SessionToken {
	t.Helper()
	return domaintoken.ReconstructSessionToken(uuid.NewString(), userID, "session-user@example.com", fixtureOrganizationName, newTestTokenHash(t), opaqueTokenCreatedAt, opaqueTokenExpiresAt, nil)
}

func Test_RefreshTokenRepository_FindByTokenHash_shouldReturnSavedToken_whenTokenExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewRefreshTokenRepository(tx)
	token := newTestRefreshToken(t, setupOpaqueTokenUser(ctx, t, tx, "opaque-rt-find"))
	require.NoError(t, repo.Save(ctx, token))

	// when
	found, err := repo.FindByTokenHash(ctx, string(token.TokenHash()))

	// then
	require.NoError(t, err)
	assert.Equal(t, token, found)
}

func Test_RefreshTokenRepository_FindByTokenHash_shouldReturnErrTokenNotFound_whenHashDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewRefreshTokenRepository(tx)

	// when
	_, err := repo.FindByTokenHash(ctx, string(newTestTokenHash(t)))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_RefreshTokenRepository_FindByTokenHash_shouldReturnErrTokenNotFound_whenSavedAsSessionToken(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	sessionToken := newTestSessionToken(t, setupOpaqueTokenUser(ctx, t, tx, "opaque-rt-cross"))
	require.NoError(t, gateway.NewSessionTokenRepository(tx).Save(ctx, sessionToken))
	repo := gateway.NewRefreshTokenRepository(tx)

	// when
	_, err := repo.FindByTokenHash(ctx, string(sessionToken.TokenHash()))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_RefreshTokenRepository_Save_shouldUpdateRevokedAt_whenTokenIsRevoked(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewRefreshTokenRepository(tx)
	token := newTestRefreshToken(t, setupOpaqueTokenUser(ctx, t, tx, "opaque-rt-revoke"))
	require.NoError(t, repo.Save(ctx, token))
	token.Revoke(opaqueTokenRevokedAt)

	// when
	err := repo.Save(ctx, token)

	// then
	require.NoError(t, err)
	found, err := repo.FindByTokenHash(ctx, string(token.TokenHash()))
	require.NoError(t, err)
	assert.Equal(t, &opaqueTokenRevokedAt, found.RevokedAt())
}

func Test_SessionTokenRepository_FindByTokenHash_shouldReturnSavedToken_whenTokenExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewSessionTokenRepository(tx)
	token := newTestSessionToken(t, setupOpaqueTokenUser(ctx, t, tx, "opaque-st-find"))
	require.NoError(t, repo.Save(ctx, token))

	// when
	found, err := repo.FindByTokenHash(ctx, string(token.TokenHash()))

	// then
	require.NoError(t, err)
	assert.Equal(t, token, found)
}

func Test_SessionTokenRepository_FindByTokenHash_shouldReturnErrTokenNotFound_whenHashDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewSessionTokenRepository(tx)

	// when
	_, err := repo.FindByTokenHash(ctx, string(newTestTokenHash(t)))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_SessionTokenRepository_FindByTokenHash_shouldReturnErrTokenNotFound_whenSavedAsRefreshToken(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	refreshToken := newTestRefreshToken(t, setupOpaqueTokenUser(ctx, t, tx, "opaque-st-cross"))
	require.NoError(t, gateway.NewRefreshTokenRepository(tx).Save(ctx, refreshToken))
	repo := gateway.NewSessionTokenRepository(tx)

	// when
	_, err := repo.FindByTokenHash(ctx, string(refreshToken.TokenHash()))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_SessionTokenRepository_Save_shouldUpdateRevokedAt_whenTokenIsRevoked(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := gateway.NewSessionTokenRepository(tx)
	token := newTestSessionToken(t, setupOpaqueTokenUser(ctx, t, tx, "opaque-st-revoke"))
	require.NoError(t, repo.Save(ctx, token))
	token.Revoke(opaqueTokenRevokedAt)

	// when
	err := repo.Save(ctx, token)

	// then
	require.NoError(t, err)
	found, err := repo.FindByTokenHash(ctx, string(token.TokenHash()))
	require.NoError(t, err)
	assert.Equal(t, &opaqueTokenRevokedAt, found.RevokedAt())
}
