package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

const fixtureSessionTokenID = "session-token-id"

func newValidateSessionTokenInput(t *testing.T, raw string) *authservice.ValidateSessionTokenInput {
	t.Helper()
	input, err := authservice.NewValidateSessionTokenInput(raw)
	require.NoError(t, err)
	return input
}

func newValidateSessionTokenConfig(now time.Time) authusecase.UsecaseConfig {
	return authusecase.UsecaseConfig{
		SessionTokenTTLMin: 30,
		SessionMaxTTLMin:   8 * 60,
		TokenWhitelistSize: 10,
		ClockFunc:          func() time.Time { return now },
	}
}

func reconstructValidSessionToken(createdAt, expiresAt time.Time, hash domain.TokenHash, revokedAt *time.Time) *domaintoken.SessionToken {
	return domaintoken.ReconstructSessionToken(
		fixtureSessionTokenID,
		fixtureAppUserID,
		domain.LoginID(fixtureLoginID),
		fixtureOrgName,
		hash,
		createdAt,
		expiresAt,
		revokedAt,
	)
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnOutput_whenTokenIsCachedAndValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(30*time.Minute), hash, nil)

	repoMock := NewMockSessionTokenRepository(t)
	whitelistMock := NewMockWhitelistRepository(t)
	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	output, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureAppUserID, output.UserID)
	assert.Equal(t, fixtureLoginID, output.LoginID)
	assert.Equal(t, fixtureOrgName, output.OrganizationName)
	repoMock.AssertNotCalled(t, "FindByTokenHash")
	whitelistMock.AssertNotCalled(t, "FindByUserID")
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldHydrateAndCache_whenCacheMissAndWhitelistContainsToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(30*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)
	cacheMock.On("SetSessionToken", string(hash), token).Return()

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(token, nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{{ID: fixtureSessionTokenID, CreatedAt: now}}, nil)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	output, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureAppUserID, output.UserID)
	cacheMock.AssertCalled(t, "SetSessionToken", string(hash), token)
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnErrTokenNotFound_whenWhitelistDoesNotContainToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(30*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(token, nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{{ID: "other-token-id", CreatedAt: now}}, nil)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	_, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
	cacheMock.AssertNotCalled(t, "SetSessionToken")
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnError_whenRepoFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)

	repoErr := errors.New("db unavailable")
	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(nil, repoErr)

	whitelistMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	_, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, repoErr)
	whitelistMock.AssertNotCalled(t, "FindByUserID")
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnError_whenWhitelistFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(30*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(token, nil)

	findErr := errors.New("whitelist query failed")
	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, findErr)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	_, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, findErr)
	cacheMock.AssertNotCalled(t, "SetSessionToken")
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnErrTokenRevoked_whenTokenIsRevoked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	revokedAt := now.Add(-5 * time.Minute)
	token := reconstructValidSessionToken(now.Add(-1*time.Hour), now.Add(30*time.Minute), hash, &revokedAt)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)

	repoMock := NewMockSessionTokenRepository(t)
	whitelistMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	_, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, domain.ErrTokenRevoked)
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnErrSessionExpired_whenTokenIsExpired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Hour), now.Add(-1*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)

	repoMock := NewMockSessionTokenRepository(t)
	whitelistMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	_, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, domain.ErrSessionExpired)
}

func Test_ValidateSessionTokenQuery_ValidateSessionToken_shouldReturnErrSessionExpired_whenAbsoluteTTLExceeded(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	// Created 9 hours ago — SessionMaxTTLMin is 8 hours, so the token has exceeded its absolute lifetime
	// even though the sliding ExpiresAt is still in the future.
	token := reconstructValidSessionToken(now.Add(-9*time.Hour), now.Add(30*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)

	repoMock := NewMockSessionTokenRepository(t)
	whitelistMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateSessionTokenQuery(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	_, err := query.ValidateSessionToken(ctx, newValidateSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, domain.ErrSessionExpired)
}
