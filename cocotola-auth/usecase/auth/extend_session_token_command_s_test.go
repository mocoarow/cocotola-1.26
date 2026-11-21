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

func newExtendSessionTokenInput(t *testing.T, raw string) *authservice.ExtendSessionTokenInput {
	t.Helper()
	input, err := authservice.NewExtendSessionTokenInput(raw)
	require.NoError(t, err)
	return input
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldSaveExtendedExpiry_whenTokenIsCachedAndValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(5*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)
	cacheMock.On("SetSessionToken", string(hash), token).Return()

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, token).Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.NoError(t, err)
	assert.Equal(t, now.Add(30*time.Minute), token.ExpiresAt())
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldCacheToken_whenTokenIsSaved(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(5*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)
	cacheMock.On("SetSessionToken", string(hash), token).Return()

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, token).Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.NoError(t, err)
	cacheMock.AssertCalled(t, "SetSessionToken", string(hash), token)
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldSaveExtendedExpiry_whenCacheMissAndWhitelistContainsToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(5*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)
	cacheMock.On("SetSessionToken", string(hash), token).Return()

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(token, nil)
	repoMock.On("Save", mock.Anything, token).Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return(newTestWhitelist(t, domaintoken.WhitelistEntry{ID: fixtureSessionTokenID, CreatedAt: now}), nil)

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.NoError(t, err)
	assert.Equal(t, now.Add(30*time.Minute), token.ExpiresAt())
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldReturnErrTokenNotFound_whenWhitelistDoesNotContainToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(5*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(token, nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return(newTestWhitelist(t, domaintoken.WhitelistEntry{ID: "other-token-id", CreatedAt: now}), nil)

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldReturnError_whenRepoFindFails(t *testing.T) {
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

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldReturnError_whenWhitelistFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(5*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(nil, false)

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, string(hash)).Return(token, nil)

	findErr := errors.New("whitelist query failed")
	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, findErr)

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, findErr)
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldReturnErrTokenRevoked_whenTokenIsRevoked(t *testing.T) {
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

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, domain.ErrTokenRevoked)
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldReturnErrSessionExpired_whenTokenIsNoLongerUsable(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		createdAt time.Time
		expiresAt time.Time
	}{
		{name: "sliding expiry passed", createdAt: now.Add(-1 * time.Hour), expiresAt: now.Add(-1 * time.Minute)},
		{name: "absolute TTL exceeded", createdAt: now.Add(-9 * time.Hour), expiresAt: now.Add(30 * time.Minute)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			raw := "raw-session"
			hash := domaintoken.HashToken(raw)
			token := reconstructValidSessionToken(tt.createdAt, tt.expiresAt, hash, nil)

			cacheMock := NewMockTokenCache(t)
			cacheMock.On("GetSessionToken", string(hash)).Return(token, true)

			repoMock := NewMockSessionTokenRepository(t)
			whitelistMock := NewMockWhitelistRepository(t)

			command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

			// when
			err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

			// then
			require.ErrorIs(t, err, domain.ErrSessionExpired)
		})
	}
}

func Test_ExtendSessionTokenCommand_ExtendSessionToken_shouldNotCacheToken_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	raw := "raw-session"
	hash := domaintoken.HashToken(raw)
	token := reconstructValidSessionToken(now.Add(-1*time.Minute), now.Add(5*time.Minute), hash, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", string(hash)).Return(token, true)

	saveErr := errors.New("save failed")
	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, token).Return(saveErr)

	whitelistMock := NewMockWhitelistRepository(t)

	command := authusecase.NewExtendSessionTokenCommand(repoMock, whitelistMock, cacheMock, newValidateSessionTokenConfig(now))

	// when
	err := command.ExtendSessionToken(ctx, newExtendSessionTokenInput(t, raw))

	// then
	require.ErrorIs(t, err, saveErr)
	cacheMock.AssertNotCalled(t, "SetSessionToken")
}
