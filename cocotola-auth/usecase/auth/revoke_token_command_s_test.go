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
	revokeFixtureJTI           = "access-token-jti"
	revokeFixtureOpaqueInput   = "raw-refresh-value"
	revokeFixtureRefreshID     = "refresh-token-id"
	revokeFixtureOtherTokenID  = "other-token-id"
	revokeFixtureActiveID1     = "active-1"
	revokeFixtureActiveID2     = "active-2"
	revokeFixtureRevokedID     = "already-revoked"
	revokeFixtureWhitelistSize = 10
)

var (
	revokeFixtureNow       = time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	errRevokeFixtureNotJWT = errors.New("not a jwt")
	errRevokeFixtureRepo   = errors.New("repository failure")
)

type revokeTokenMocks struct {
	jwt              *MockJWTManager
	refreshRepo      *MockRefreshTokenRepository
	accessRepo       *MockAccessTokenRepository
	refreshWhitelist *MockWhitelistRepository
	accessWhitelist  *MockWhitelistRepository
	cache            *MockTokenCache
}

func newRevokeTokenMocks(t *testing.T) revokeTokenMocks {
	t.Helper()
	return revokeTokenMocks{
		jwt:              NewMockJWTManager(t),
		refreshRepo:      NewMockRefreshTokenRepository(t),
		accessRepo:       NewMockAccessTokenRepository(t),
		refreshWhitelist: NewMockWhitelistRepository(t),
		accessWhitelist:  NewMockWhitelistRepository(t),
		cache:            NewMockTokenCache(t),
	}
}

func (m revokeTokenMocks) newCommand() *authusecase.RevokeTokenCommand {
	return authusecase.NewRevokeTokenCommand(
		m.refreshRepo,
		m.accessRepo,
		m.refreshWhitelist,
		m.accessWhitelist,
		m.jwt,
		m.cache,
		authusecase.UsecaseConfig{
			TokenWhitelistSize: revokeFixtureWhitelistSize,
			ClockFunc:          func() time.Time { return revokeFixtureNow },
		},
	)
}

func newRevokeFixtureAccessToken(id string, revokedAt *time.Time) *domaintoken.AccessToken {
	return domaintoken.ReconstructAccessToken(id, revokeFixtureRefreshID, fixtureAppUserID, domain.LoginID(fixtureLoginID), fixtureOrgName,
		revokeFixtureNow.Add(-time.Hour), revokeFixtureNow.Add(time.Hour), revokedAt)
}

func newRevokeFixtureRefreshToken() *domaintoken.RefreshToken {
	return domaintoken.ReconstructRefreshToken(revokeFixtureRefreshID, fixtureAppUserID, domain.LoginID(fixtureLoginID), fixtureOrgName,
		domaintoken.HashToken(revokeFixtureOpaqueInput), revokeFixtureNow.Add(-time.Hour), revokeFixtureNow.Add(24*time.Hour), nil)
}

func newRevokeFixtureWhitelistEntries(ids ...string) []domaintoken.WhitelistEntry {
	entries := make([]domaintoken.WhitelistEntry, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, domaintoken.WhitelistEntry{ID: id, CreatedAt: revokeFixtureNow.Add(-time.Hour)})
	}
	return entries
}

func newRevokeTokenInput(t *testing.T, token string) *authservice.RevokeTokenInput {
	t.Helper()
	input, err := authservice.NewRevokeTokenInput(token)
	require.NoError(t, err)
	return input
}

type savedWhitelist struct {
	entries []domaintoken.WhitelistEntry
}

func expectWhitelistSaved(t *testing.T, repo *MockWhitelistRepository, existing []domaintoken.WhitelistEntry) *savedWhitelist {
	t.Helper()
	saved := &savedWhitelist{}
	repo.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).
		Run(func(args mock.Arguments) {
			wl, ok := args.Get(1).(*domaintoken.Whitelist)
			require.True(t, ok, "second arg must be *Whitelist")
			saved.entries = wl.Entries()
		}).
		Return(nil)
	return saved
}

type savedAccessTokens struct {
	tokens []domaintoken.AccessToken
}

func (s *savedAccessTokens) ids() []string {
	ids := make([]string, 0, len(s.tokens))
	for i := range s.tokens {
		ids = append(ids, s.tokens[i].ID())
	}
	return ids
}

func (s *savedAccessTokens) revokedAts() []time.Time {
	revokedAts := make([]time.Time, 0, len(s.tokens))
	for i := range s.tokens {
		var revokedAt time.Time
		if s.tokens[i].RevokedAt() != nil {
			revokedAt = *s.tokens[i].RevokedAt()
		}
		revokedAts = append(revokedAts, revokedAt)
	}
	return revokedAts
}

func expectAccessTokensSaved(t *testing.T, repo *MockAccessTokenRepository) *savedAccessTokens {
	t.Helper()
	saved := &savedAccessTokens{}
	repo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.AccessToken)
			require.True(t, ok, "second arg must be *AccessToken")
			saved.tokens = append(saved.tokens, *tk)
		}).
		Return(nil)
	return saved
}

func givenJWTAccessTokenFound(m revokeTokenMocks) {
	m.jwt.On("ParseAccessToken", fixtureJWT).Return(nil, revokeFixtureJTI, nil)
	m.accessRepo.On("FindByID", mock.Anything, revokeFixtureJTI).Return(newRevokeFixtureAccessToken(revokeFixtureJTI, nil), nil)
}

func givenJWTAccessTokenRevocable(m revokeTokenMocks) {
	givenJWTAccessTokenFound(m)
	m.cache.On("DeleteAccessToken", revokeFixtureJTI).Return()
}

type savedRefreshToken struct {
	token *domaintoken.RefreshToken
}

func givenOpaqueRefreshTokenFound(m revokeTokenMocks) {
	m.jwt.On("ParseAccessToken", revokeFixtureOpaqueInput).Return(nil, "", errRevokeFixtureNotJWT)
	m.refreshRepo.On("FindByTokenHash", mock.Anything, string(domaintoken.HashToken(revokeFixtureOpaqueInput))).
		Return(newRevokeFixtureRefreshToken(), nil)
}

func givenOpaqueRefreshToken(t *testing.T, m revokeTokenMocks, linked []domaintoken.AccessToken) *savedRefreshToken {
	t.Helper()
	saved := &savedRefreshToken{}
	givenOpaqueRefreshTokenFound(m)
	m.accessRepo.On("FindByRefreshTokenID", mock.Anything, revokeFixtureRefreshID).Return(linked, nil)
	m.refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).
		Run(func(args mock.Arguments) {
			tk, ok := args.Get(1).(*domaintoken.RefreshToken)
			require.True(t, ok, "second arg must be *RefreshToken")
			saved.token = tk
		}).
		Return(nil)
	return saved
}

func newRevokeFixtureLinkedAccessTokens() []domaintoken.AccessToken {
	revokedAt := revokeFixtureNow.Add(-10 * time.Minute)
	return []domaintoken.AccessToken{
		*newRevokeFixtureAccessToken(revokeFixtureActiveID1, nil),
		*newRevokeFixtureAccessToken(revokeFixtureRevokedID, &revokedAt),
		*newRevokeFixtureAccessToken(revokeFixtureActiveID2, nil),
	}
}

func givenLinkedAccessTokensRevocable(m revokeTokenMocks) {
	m.cache.On("DeleteAccessToken", revokeFixtureActiveID1).Return()
	m.cache.On("DeleteAccessToken", revokeFixtureActiveID2).Return()
}

func Test_RevokeTokenCommand_RevokeToken_shouldSaveAccessTokenRevokedAtNow_whenTokenIsJWT(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenJWTAccessTokenRevocable(m)
	saved := expectAccessTokensSaved(t, m.accessRepo)
	expectWhitelistSaved(t, m.accessWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureJTI))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, fixtureJWT))

	// then
	require.NoError(t, err)
	require.Len(t, saved.tokens, 1)
	require.NotNil(t, saved.tokens[0].RevokedAt())
	assert.Equal(t, revokeFixtureNow, *saved.tokens[0].RevokedAt())
}

func Test_RevokeTokenCommand_RevokeToken_shouldRemoveOnlyJTIFromAccessWhitelist_whenTokenIsJWT(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenJWTAccessTokenRevocable(m)
	expectAccessTokensSaved(t, m.accessRepo)
	saved := expectWhitelistSaved(t, m.accessWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureOtherTokenID, revokeFixtureJTI))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, fixtureJWT))

	// then
	require.NoError(t, err)
	assert.Equal(t, newRevokeFixtureWhitelistEntries(revokeFixtureOtherTokenID), saved.entries)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnErrTokenNotFound_whenAccessTokenIsNotFound(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	m.jwt.On("ParseAccessToken", fixtureJWT).Return(nil, revokeFixtureJTI, nil)
	m.accessRepo.On("FindByID", mock.Anything, revokeFixtureJTI).Return(nil, domain.ErrTokenNotFound)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, fixtureJWT))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_RevokeTokenCommand_RevokeToken_shouldSaveRefreshTokenRevokedAtNow_whenTokenIsOpaque(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	saved := givenOpaqueRefreshToken(t, m, []domaintoken.AccessToken{})
	expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.NoError(t, err)
	require.NotNil(t, saved.token)
	require.NotNil(t, saved.token.RevokedAt())
	assert.Equal(t, revokeFixtureNow, *saved.token.RevokedAt())
}

func Test_RevokeTokenCommand_RevokeToken_shouldSaveOnlyActiveLinkedAccessTokens_whenTokenIsOpaque(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, newRevokeFixtureLinkedAccessTokens())
	givenLinkedAccessTokensRevocable(m)
	saved := expectAccessTokensSaved(t, m.accessRepo)
	expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID))
	expectWhitelistSaved(t, m.accessWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureActiveID1, revokeFixtureActiveID2))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{revokeFixtureActiveID1, revokeFixtureActiveID2}, saved.ids())
}

func Test_RevokeTokenCommand_RevokeToken_shouldSaveLinkedAccessTokensRevokedAtNow_whenTokenIsOpaque(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, newRevokeFixtureLinkedAccessTokens())
	givenLinkedAccessTokensRevocable(m)
	saved := expectAccessTokensSaved(t, m.accessRepo)
	expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID))
	expectWhitelistSaved(t, m.accessWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureActiveID1, revokeFixtureActiveID2))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.NoError(t, err)
	assert.Equal(t, []time.Time{revokeFixtureNow, revokeFixtureNow}, saved.revokedAts())
}

func Test_RevokeTokenCommand_RevokeToken_shouldRemoveOnlyRefreshTokenFromRefreshWhitelist_whenTokenIsOpaque(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, []domaintoken.AccessToken{})
	saved := expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID, revokeFixtureOtherTokenID))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.NoError(t, err)
	assert.Equal(t, newRevokeFixtureWhitelistEntries(revokeFixtureOtherTokenID), saved.entries)
}

func Test_RevokeTokenCommand_RevokeToken_shouldRemoveRevokedAccessTokensFromAccessWhitelist_whenTokenIsOpaque(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, newRevokeFixtureLinkedAccessTokens())
	givenLinkedAccessTokensRevocable(m)
	expectAccessTokensSaved(t, m.accessRepo)
	expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID))
	saved := expectWhitelistSaved(t, m.accessWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureActiveID1, revokeFixtureOtherTokenID, revokeFixtureActiveID2))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.NoError(t, err)
	assert.Equal(t, newRevokeFixtureWhitelistEntries(revokeFixtureOtherTokenID), saved.entries)
}

func Test_RevokeTokenCommand_RevokeToken_shouldNotTouchAccessWhitelist_whenAllLinkedAccessTokensAreRevoked(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	revokedAt := revokeFixtureNow.Add(-10 * time.Minute)
	givenOpaqueRefreshToken(t, m, []domaintoken.AccessToken{*newRevokeFixtureAccessToken(revokeFixtureRevokedID, &revokedAt)})
	expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID))

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.NoError(t, err)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnErrTokenNotFound_whenRefreshTokenIsNotFound(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	m.jwt.On("ParseAccessToken", revokeFixtureOpaqueInput).Return(nil, "", errRevokeFixtureNotJWT)
	m.refreshRepo.On("FindByTokenHash", mock.Anything, string(domaintoken.HashToken(revokeFixtureOpaqueInput))).
		Return(nil, domain.ErrTokenNotFound)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenAccessTokenSaveFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenJWTAccessTokenFound(m)
	m.accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).Return(errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, fixtureJWT))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenLinkedAccessTokensFindFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshTokenFound(m)
	m.accessRepo.On("FindByRefreshTokenID", mock.Anything, revokeFixtureRefreshID).Return(nil, errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenRefreshTokenSaveFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshTokenFound(m)
	m.accessRepo.On("FindByRefreshTokenID", mock.Anything, revokeFixtureRefreshID).Return(newRevokeFixtureLinkedAccessTokens(), nil)
	m.refreshRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.RefreshToken")).Return(errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenLinkedAccessTokenSaveFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, newRevokeFixtureLinkedAccessTokens())
	m.accessRepo.On("Save", mock.Anything, mock.AnythingOfType("*token.AccessToken")).Return(errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenAccessWhitelistFindFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenJWTAccessTokenRevocable(m)
	expectAccessTokensSaved(t, m.accessRepo)
	m.accessWhitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, fixtureJWT))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenAccessWhitelistSaveFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenJWTAccessTokenRevocable(m)
	expectAccessTokensSaved(t, m.accessRepo)
	m.accessWhitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(newRevokeFixtureWhitelistEntries(revokeFixtureJTI), nil)
	m.accessWhitelist.On("Save", mock.Anything, mock.AnythingOfType("*token.Whitelist")).Return(errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, fixtureJWT))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenRefreshWhitelistFindFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, newRevokeFixtureLinkedAccessTokens())
	givenLinkedAccessTokensRevocable(m)
	expectAccessTokensSaved(t, m.accessRepo)
	m.refreshWhitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}

func Test_RevokeTokenCommand_RevokeToken_shouldReturnError_whenLinkedAccessWhitelistFindFails(t *testing.T) {
	t.Parallel()

	// given
	m := newRevokeTokenMocks(t)
	givenOpaqueRefreshToken(t, m, newRevokeFixtureLinkedAccessTokens())
	givenLinkedAccessTokensRevocable(m)
	expectAccessTokensSaved(t, m.accessRepo)
	expectWhitelistSaved(t, m.refreshWhitelist, newRevokeFixtureWhitelistEntries(revokeFixtureRefreshID))
	m.accessWhitelist.On("FindByUserID", mock.Anything, fixtureAppUserID).Return(nil, errRevokeFixtureRepo)

	// when
	err := m.newCommand().RevokeToken(context.Background(), newRevokeTokenInput(t, revokeFixtureOpaqueInput))

	// then
	require.ErrorIs(t, err, errRevokeFixtureRepo)
}
