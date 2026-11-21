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

const (
	refreshFixtureOpaqueInput = "raw-refresh-value"
	refreshFixtureTokenID     = "refresh-token-id"
)

var refreshFixtureNow = time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

type refreshAccessTokenMocks struct {
	refreshRepo *MockRefreshTokenRepository
	accessRepo  *MockAccessTokenRepository
	whitelist   *MockWhitelistRepository
	jwt         *MockJWTManager
	cache       *MockTokenCache
}

func newRefreshAccessTokenMocks(t *testing.T) refreshAccessTokenMocks {
	t.Helper()
	refreshToken := domaintoken.ReconstructRefreshToken(refreshFixtureTokenID, fixtureAppUserID, domain.LoginID(fixtureLoginID), fixtureOrgName,
		domaintoken.HashToken(refreshFixtureOpaqueInput), refreshFixtureNow.Add(-time.Hour), refreshFixtureNow.Add(time.Hour), nil)
	m := refreshAccessTokenMocks{
		refreshRepo: NewMockRefreshTokenRepository(t),
		accessRepo:  NewMockAccessTokenRepository(t),
		whitelist:   NewMockWhitelistRepository(t),
		jwt:         NewMockJWTManager(t),
		cache:       NewMockTokenCache(t),
	}
	m.refreshRepo.On("FindByTokenHash", mock.Anything, string(domaintoken.HashToken(refreshFixtureOpaqueInput))).Return(refreshToken, nil)
	expectJWTCreated(t, m.jwt)
	m.cache.On("SetAccessToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.AccessToken")).Return()
	return m
}

func (m refreshAccessTokenMocks) captureSavedAccessToken(t *testing.T) *domaintoken.AccessToken {
	t.Helper()
	saved := new(domaintoken.AccessToken)
	m.accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.AccessToken)
			require.True(t, ok, "second arg must be *AccessToken")
			*saved = *tk
		}).
		Return(nil)
	return saved
}

func (m refreshAccessTokenMocks) captureSavedWhitelist(t *testing.T) *domaintoken.Whitelist {
	t.Helper()
	saved := new(domaintoken.Whitelist)
	m.whitelist.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).
		Run(func(args mock.Arguments) {
			w, ok := args.Get(1).(*domaintoken.Whitelist)
			require.True(t, ok, "second arg must be *Whitelist")
			*saved = *w
		}).
		Return(nil)
	return saved
}

func (m refreshAccessTokenMocks) newCommand() *authusecase.RefreshAccessTokenCommand {
	return authusecase.NewRefreshAccessTokenCommand(m.refreshRepo, m.accessRepo, m.whitelist, m.jwt, m.cache, authusecase.UsecaseConfig{
		AccessTokenTTLMin:  15,
		TokenWhitelistSize: fixtureWhitelistSize,
		ClockFunc:          func() time.Time { return refreshFixtureNow },
	})
}

func newRefreshAccessTokenInput(t *testing.T) *authservice.RefreshAccessTokenInput {
	t.Helper()
	input, err := authservice.NewRefreshAccessTokenInput(refreshFixtureOpaqueInput)
	require.NoError(t, err)
	return input
}

func Test_RefreshAccessTokenCommand_RefreshAccessToken_shouldSaveWhitelistContainingIssuedAccessToken_whenRefreshTokenIsValid(t *testing.T) {
	t.Parallel()

	// given
	m := newRefreshAccessTokenMocks(t)
	m.whitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(newTestWhitelist(t), nil)
	savedAccess := m.captureSavedAccessToken(t)
	saved := m.captureSavedWhitelist(t)

	// when
	_, err := m.newCommand().RefreshAccessToken(context.Background(), newRefreshAccessTokenInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []domaintoken.WhitelistEntry{{ID: savedAccess.ID(), CreatedAt: refreshFixtureNow}}, saved.Entries())
}

func Test_RefreshAccessTokenCommand_RefreshAccessToken_shouldEvictOldestEntry_whenWhitelistIsFull(t *testing.T) {
	t.Parallel()

	// given
	m := newRefreshAccessTokenMocks(t)
	m.whitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(fullTestWhitelist(t, refreshFixtureNow), nil)
	savedAccess := m.captureSavedAccessToken(t)
	saved := m.captureSavedWhitelist(t)

	// when
	_, err := m.newCommand().RefreshAccessToken(context.Background(), newRefreshAccessTokenInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"newer", savedAccess.ID()}, whitelistEntryIDs(saved))
}

func Test_RefreshAccessTokenCommand_RefreshAccessToken_shouldReturnError_whenWhitelistSaveFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRefreshAccessTokenMocks(t)
	m.whitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(newTestWhitelist(t), nil)
	saveErr := errors.New("whitelist save failed")
	m.accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).Return(nil)
	m.whitelist.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(saveErr)

	// when
	_, err := m.newCommand().RefreshAccessToken(context.Background(), newRefreshAccessTokenInput(t))

	// then
	require.ErrorIs(t, err, saveErr)
}
