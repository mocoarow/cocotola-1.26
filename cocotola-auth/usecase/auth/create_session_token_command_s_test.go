package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

const (
	fixtureLoginID = "user1@example.com"
	fixtureOrgName = "org1"
)

func newCreateSessionTokenInput(t *testing.T) *authservice.CreateSessionTokenInput {
	t.Helper()
	input, err := authservice.NewCreateSessionTokenInput(fixtureAppUserID, fixtureLoginID, fixtureOrgName)
	require.NoError(t, err)
	return input
}

func newCreateSessionTokenConfig(now time.Time) authusecase.UsecaseConfig {
	return authusecase.UsecaseConfig{
		SessionTokenTTLMin: 30,
		TokenWhitelistSize: 10,
		ClockFunc:          func() time.Time { return now },
	}
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldReturnRawToken_whenAllSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	repoMock := NewMockSessionTokenRepository(t)
	var savedToken *domaintoken.SessionToken
	repoMock.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.SessionToken)
			require.True(t, ok, "second arg must be *SessionToken")
			savedToken = tk
		}).
		Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	whitelistMock.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetSessionToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.SessionToken")).Return()

	cmd := authusecase.NewCreateSessionTokenCommand(
		repoMock,
		whitelistMock,
		cacheMock,
		newCreateSessionTokenConfig(now),
	)

	// when
	output, err := cmd.CreateSessionToken(ctx, newCreateSessionTokenInput(t))

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, output.RawToken)
	require.NotNil(t, savedToken)
	assert.Equal(t, fixtureAppUserID, savedToken.UserID())
	assert.Equal(t, fixtureOrgName, savedToken.OrganizationName())
	assert.Equal(t, now, savedToken.CreatedAt())
	assert.Equal(t, now.Add(30*time.Minute), savedToken.ExpiresAt())
	assert.Equal(t, domaintoken.HashToken(output.RawToken), savedToken.TokenHash())
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldCacheTokenByHash_whenSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	whitelistMock.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	var cachedHash string
	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetSessionToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.SessionToken")).
		Run(func(args mock.Arguments) {
			h, ok := args.Get(0).(string)
			require.True(t, ok, "first arg must be string")
			cachedHash = h
		}).
		Return()

	cmd := authusecase.NewCreateSessionTokenCommand(
		repoMock,
		whitelistMock,
		cacheMock,
		newCreateSessionTokenConfig(now),
	)

	// when
	output, err := cmd.CreateSessionToken(ctx, newCreateSessionTokenInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, string(domaintoken.HashToken(output.RawToken)), cachedHash)
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldReturnError_whenSessionTokenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	saveErr := errors.New("write failed")
	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).Return(saveErr)

	whitelistMock := NewMockWhitelistRepository(t)
	cacheMock := NewMockTokenCache(t)

	cmd := authusecase.NewCreateSessionTokenCommand(
		repoMock,
		whitelistMock,
		cacheMock,
		newCreateSessionTokenConfig(now),
	)

	// when
	_, err := cmd.CreateSessionToken(ctx, newCreateSessionTokenInput(t))

	// then
	require.ErrorIs(t, err, saveErr)
	cacheMock.AssertNotCalled(t, "SetSessionToken")
	whitelistMock.AssertNotCalled(t, "FindByUserID")
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldReturnError_whenWhitelistFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	findErr := errors.New("whitelist query failed")
	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, findErr)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetSessionToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.SessionToken")).Return()

	cmd := authusecase.NewCreateSessionTokenCommand(
		repoMock,
		whitelistMock,
		cacheMock,
		newCreateSessionTokenConfig(now),
	)

	// when
	_, err := cmd.CreateSessionToken(ctx, newCreateSessionTokenInput(t))

	// then
	require.ErrorIs(t, err, findErr)
	whitelistMock.AssertNotCalled(t, "Save")
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldReturnError_whenWhitelistSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	saveErr := errors.New("whitelist save failed")
	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).Return(nil)

	whitelistMock := NewMockWhitelistRepository(t)
	whitelistMock.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	whitelistMock.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(saveErr)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetSessionToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.SessionToken")).Return()

	cmd := authusecase.NewCreateSessionTokenCommand(
		repoMock,
		whitelistMock,
		cacheMock,
		newCreateSessionTokenConfig(now),
	)

	// when
	_, err := cmd.CreateSessionToken(ctx, newCreateSessionTokenInput(t))

	// then
	require.ErrorIs(t, err, saveErr)
}
