package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"
)

var whitelistUpdateFixtureNow = time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

type whitelistUpdateMocks struct {
	repo      *MockSessionTokenRepository
	whitelist *MockWhitelistRepository
	cache     *MockTokenCache
}

func newWhitelistUpdateMocks(t *testing.T) whitelistUpdateMocks {
	t.Helper()
	m := whitelistUpdateMocks{
		repo:      NewMockSessionTokenRepository(t),
		whitelist: NewMockWhitelistRepository(t),
		cache:     NewMockTokenCache(t),
	}
	m.cache.On("SetSessionToken", mock.AnythingOfType("string"), mock.AnythingOfType("*token.SessionToken")).Return()
	return m
}

func (m whitelistUpdateMocks) givenSessionTokenSaved() {
	m.repo.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).Return(nil)
}

func (m whitelistUpdateMocks) givenFreshWhitelistOnEveryFind(t *testing.T) {
	t.Helper()
	m.whitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).
		Return(func(context.Context, domain.AppUserID) *domaintoken.Whitelist { return newTestWhitelist(t) }, nil)
}

func (m whitelistUpdateMocks) newCommand() *authusecase.CreateSessionTokenCommand {
	return authusecase.NewCreateSessionTokenCommand(m.repo, m.whitelist, m.cache, newCreateSessionTokenConfig(whitelistUpdateFixtureNow))
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldSucceed_whenWhitelistSaveConflictsOnce(t *testing.T) {
	t.Parallel()

	// given
	m := newWhitelistUpdateMocks(t)
	m.givenSessionTokenSaved()
	m.givenFreshWhitelistOnEveryFind(t)
	m.whitelist.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	m.whitelist.On("Save", mock.Anything, mock.Anything).Return(nil)

	// when
	_, err := m.newCommand().CreateSessionToken(context.Background(), newCreateSessionTokenInput(t))

	// then
	require.NoError(t, err)
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldRereadWhitelist_whenWhitelistSaveConflicts(t *testing.T) {
	t.Parallel()

	// given
	m := newWhitelistUpdateMocks(t)
	m.givenSessionTokenSaved()
	m.givenFreshWhitelistOnEveryFind(t)
	m.whitelist.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	m.whitelist.On("Save", mock.Anything, mock.Anything).Return(nil)

	// when
	_, _ = m.newCommand().CreateSessionToken(context.Background(), newCreateSessionTokenInput(t))

	// then
	m.whitelist.AssertNumberOfCalls(t, "FindByUserID", 2)
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldReturnErrConcurrentModification_whenWhitelistSaveAlwaysConflicts(t *testing.T) {
	t.Parallel()

	// given
	m := newWhitelistUpdateMocks(t)
	m.givenSessionTokenSaved()
	m.givenFreshWhitelistOnEveryFind(t)
	m.whitelist.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	// when
	_, err := m.newCommand().CreateSessionToken(context.Background(), newCreateSessionTokenInput(t))

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldStopRetrying_whenWhitelistSaveAttemptsAreExhausted(t *testing.T) {
	t.Parallel()

	// given
	m := newWhitelistUpdateMocks(t)
	m.givenSessionTokenSaved()
	m.givenFreshWhitelistOnEveryFind(t)
	m.whitelist.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	// when
	_, _ = m.newCommand().CreateSessionToken(context.Background(), newCreateSessionTokenInput(t))

	// then
	m.whitelist.AssertNumberOfCalls(t, "Save", occ.MaxAttempts)
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldEvictOldestEntry_whenWhitelistIsFull(t *testing.T) {
	t.Parallel()

	// given
	m := newWhitelistUpdateMocks(t)
	var savedToken *domaintoken.SessionToken
	m.repo.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.SessionToken)
			require.True(t, ok, "second arg must be *SessionToken")
			savedToken = tk
		}).
		Return(nil)
	m.whitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(fullTestWhitelist(t, whitelistUpdateFixtureNow), nil)
	var saved *domaintoken.Whitelist
	m.whitelist.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).
		Run(func(args mock.Arguments) {
			w, ok := args.Get(1).(*domaintoken.Whitelist)
			require.True(t, ok, "second arg must be *Whitelist")
			saved = w
		}).
		Return(nil)

	// when
	_, err := m.newCommand().CreateSessionToken(context.Background(), newCreateSessionTokenInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"newer", savedToken.ID()}, whitelistEntryIDs(saved))
}

func Test_CreateSessionTokenCommand_CreateSessionToken_shouldSaveWhitelistContainingSavedToken_whenSucceeds(t *testing.T) {
	t.Parallel()

	// given
	m := newWhitelistUpdateMocks(t)
	var savedToken *domaintoken.SessionToken
	m.repo.On("Save", mock.Anything, mock.AnythingOfType("*token.SessionToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.SessionToken)
			require.True(t, ok, "second arg must be *SessionToken")
			savedToken = tk
		}).
		Return(nil)
	m.whitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(newTestWhitelist(t), nil)
	var saved *domaintoken.Whitelist
	m.whitelist.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).
		Run(func(args mock.Arguments) {
			w, ok := args.Get(1).(*domaintoken.Whitelist)
			require.True(t, ok, "second arg must be *Whitelist")
			saved = w
		}).
		Return(nil)

	// when
	_, err := m.newCommand().CreateSessionToken(context.Background(), newCreateSessionTokenInput(t))

	// then
	require.NoError(t, err)
	require.NotNil(t, savedToken)
	assert.Equal(t, []domaintoken.WhitelistEntry{{ID: savedToken.ID(), CreatedAt: whitelistUpdateFixtureNow}}, saved.Entries())
}
