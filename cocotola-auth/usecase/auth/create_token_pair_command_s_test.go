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

const fixtureJWT = "header.payload.signature"

func newCreateTokenPairInput(t *testing.T) *authservice.CreateTokenPairInput {
	t.Helper()
	input, err := authservice.NewCreateTokenPairInput(fixtureAppUserID, fixtureLoginID, fixtureOrgName)
	require.NoError(t, err)
	return input
}

func newCreateTokenPairConfig(now time.Time) authusecase.UsecaseConfig {
	return authusecase.UsecaseConfig{
		AccessTokenTTLMin:  15,
		RefreshTokenTTLMin: 60 * 24,
		TokenWhitelistSize: 10,
		ClockFunc:          func() time.Time { return now },
	}
}

func expectJWTCreated(t *testing.T, jwtMock *MockJWTManager) {
	t.Helper()
	jwtMock.On(
		"CreateAccessToken",
		fixtureLoginID,
		fixtureAppUserID,
		fixtureOrgName,
		mock.AnythingOfType("string"),
	).Return(fixtureJWT, nil)
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldReturnOutput_whenAllSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	var savedRefresh *domaintoken.RefreshToken
	refreshRepo := NewMockRefreshTokenRepository(t)
	refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.RefreshToken)
			require.True(t, ok, "second arg must be *RefreshToken")
			savedRefresh = tk
		}).
		Return(nil)

	var savedAccess *domaintoken.AccessToken
	accessRepo := NewMockAccessTokenRepository(t)
	accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.AccessToken)
			require.True(t, ok, "second arg must be *AccessToken")
			savedAccess = tk
		}).
		Return(nil)

	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	refreshWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	refreshWhitelistRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	accessWhitelistRepo := NewMockWhitelistRepository(t)
	accessWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	accessWhitelistRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	jwtMock := NewMockJWTManager(t)
	expectJWTCreated(t, jwtMock)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetAccessToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.AccessToken")).Return()

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	output, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureJWT, output.AccessToken)
	assert.NotEmpty(t, output.RefreshToken)

	require.NotNil(t, savedRefresh)
	assert.Equal(t, fixtureAppUserID, savedRefresh.UserID())
	assert.Equal(t, domain.LoginID(fixtureLoginID), savedRefresh.LoginID())
	assert.Equal(t, fixtureOrgName, savedRefresh.OrganizationName())
	assert.Equal(t, now, savedRefresh.CreatedAt())
	assert.Equal(t, now.Add(24*time.Hour), savedRefresh.ExpiresAt())
	assert.Equal(t, domaintoken.HashToken(output.RefreshToken), savedRefresh.TokenHash())

	require.NotNil(t, savedAccess)
	assert.Equal(t, fixtureAppUserID, savedAccess.UserID())
	assert.Equal(t, savedRefresh.ID(), savedAccess.RefreshTokenID())
	assert.Equal(t, now, savedAccess.CreatedAt())
	assert.Equal(t, now.Add(15*time.Minute), savedAccess.ExpiresAt())
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldCacheAccessTokenByID_whenSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	refreshRepo := NewMockRefreshTokenRepository(t)
	refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).Return(nil)

	var savedAccess *domaintoken.AccessToken
	accessRepo := NewMockAccessTokenRepository(t)
	accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.AccessToken)
			require.True(t, ok, "second arg must be *AccessToken")
			savedAccess = tk
		}).
		Return(nil)

	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	refreshWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	refreshWhitelistRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	accessWhitelistRepo := NewMockWhitelistRepository(t)
	accessWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	accessWhitelistRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	jwtMock := NewMockJWTManager(t)
	expectJWTCreated(t, jwtMock)

	var cachedJTI string
	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetAccessToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.AccessToken")).
		Run(func(args mock.Arguments) {
			jti, ok := args.Get(0).(string)
			require.True(t, ok, "first arg must be string")
			cachedJTI = jti
		}).
		Return()

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	_, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.NoError(t, err)
	require.NotNil(t, savedAccess)
	assert.Equal(t, savedAccess.ID(), cachedJTI)
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldReturnError_whenJWTCreationFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	jwtErr := errors.New("signing key missing")
	refreshRepo := NewMockRefreshTokenRepository(t)
	accessRepo := NewMockAccessTokenRepository(t)
	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	accessWhitelistRepo := NewMockWhitelistRepository(t)
	cacheMock := NewMockTokenCache(t)

	jwtMock := NewMockJWTManager(t)
	jwtMock.On(
		"CreateAccessToken",
		fixtureLoginID,
		fixtureAppUserID,
		fixtureOrgName,
		mock.AnythingOfType("string"),
	).Return("", jwtErr)

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	_, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.ErrorIs(t, err, jwtErr)
	refreshRepo.AssertNotCalled(t, "Save")
	accessRepo.AssertNotCalled(t, "Save")
	cacheMock.AssertNotCalled(t, "SetAccessToken")
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldReturnError_whenRefreshTokenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	saveErr := errors.New("refresh save failed")
	refreshRepo := NewMockRefreshTokenRepository(t)
	refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).Return(saveErr)

	accessRepo := NewMockAccessTokenRepository(t)
	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	accessWhitelistRepo := NewMockWhitelistRepository(t)
	cacheMock := NewMockTokenCache(t)

	jwtMock := NewMockJWTManager(t)
	expectJWTCreated(t, jwtMock)

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	_, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.ErrorIs(t, err, saveErr)
	accessRepo.AssertNotCalled(t, "Save")
	cacheMock.AssertNotCalled(t, "SetAccessToken")
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldReturnError_whenAccessTokenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	saveErr := errors.New("access save failed")
	refreshRepo := NewMockRefreshTokenRepository(t)
	refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).Return(nil)

	accessRepo := NewMockAccessTokenRepository(t)
	accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).Return(saveErr)

	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	accessWhitelistRepo := NewMockWhitelistRepository(t)
	cacheMock := NewMockTokenCache(t)

	jwtMock := NewMockJWTManager(t)
	expectJWTCreated(t, jwtMock)

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	_, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.ErrorIs(t, err, saveErr)
	cacheMock.AssertNotCalled(t, "SetAccessToken")
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldReturnError_whenRefreshWhitelistFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	findErr := errors.New("refresh whitelist find failed")
	refreshRepo := NewMockRefreshTokenRepository(t)
	refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).Return(nil)

	accessRepo := NewMockAccessTokenRepository(t)
	accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).Return(nil)

	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	refreshWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, findErr)

	accessWhitelistRepo := NewMockWhitelistRepository(t)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetAccessToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.AccessToken")).Return()

	jwtMock := NewMockJWTManager(t)
	expectJWTCreated(t, jwtMock)

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	_, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.ErrorIs(t, err, findErr)
	refreshWhitelistRepo.AssertNotCalled(t, "Save")
	accessWhitelistRepo.AssertNotCalled(t, "FindByUserID")
}

func Test_CreateTokenPairCommand_CreateTokenPair_shouldReturnError_whenAccessWhitelistSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// given
	saveErr := errors.New("access whitelist save failed")
	refreshRepo := NewMockRefreshTokenRepository(t)
	refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).Return(nil)

	accessRepo := NewMockAccessTokenRepository(t)
	accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).Return(nil)

	refreshWhitelistRepo := NewMockWhitelistRepository(t)
	refreshWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	refreshWhitelistRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(nil)

	accessWhitelistRepo := NewMockWhitelistRepository(t)
	accessWhitelistRepo.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return([]domaintoken.WhitelistEntry{}, nil)
	accessWhitelistRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(saveErr)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("SetAccessToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.AccessToken")).Return()

	jwtMock := NewMockJWTManager(t)
	expectJWTCreated(t, jwtMock)

	cmd := authusecase.NewCreateTokenPairCommand(
		refreshRepo,
		accessRepo,
		refreshWhitelistRepo,
		accessWhitelistRepo,
		jwtMock,
		cacheMock,
		newCreateTokenPairConfig(now),
	)

	// when
	_, err := cmd.CreateTokenPair(ctx, newCreateTokenPairInput(t))

	// then
	require.ErrorIs(t, err, saveErr)
}
