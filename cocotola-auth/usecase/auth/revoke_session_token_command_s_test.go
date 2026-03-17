package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// --- mocks ---

type mockRevokeSessionTokenRepo struct{ mock.Mock }

func (m *mockRevokeSessionTokenRepo) FindByTokenHash(ctx context.Context, hash string) (*domain.SessionToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SessionToken), args.Error(1)
}

func (m *mockRevokeSessionTokenRepo) Save(ctx context.Context, token *domain.SessionToken) error {
	return m.Called(ctx, token).Error(0)
}

type mockRevokeSessionTokenWhitelistRepo struct{ mock.Mock }

func (m *mockRevokeSessionTokenWhitelistRepo) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.WhitelistEntry), args.Error(1)
}

func (m *mockRevokeSessionTokenWhitelistRepo) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return m.Called(ctx, whitelist).Error(0)
}

type mockRevokeSessionTokenCache struct{ mock.Mock }

func (m *mockRevokeSessionTokenCache) GetSessionToken(hash string) (*domain.SessionToken, bool) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*domain.SessionToken), args.Bool(1)
}

func (m *mockRevokeSessionTokenCache) DeleteSessionToken(hash string) {
	m.Called(hash)
}

// --- tests ---

func Test_RevokeSessionTokenCommand_RevokeSessionToken_shouldRevokeToken_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	rawToken := "raw-session-token"
	hash := string(domain.HashToken(rawToken))
	tokenID := "session-token-id"
	userID := 1

	sessionToken := domain.ReconstructSessionToken(tokenID, userID, "user1", "org1", domain.TokenHash(hash), now, now.Add(30*time.Minute), nil)

	repoMock := &mockRevokeSessionTokenRepo{}
	repoMock.On("FindByTokenHash", mock.Anything, hash).Return(sessionToken, nil)
	repoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	whitelistRepoMock := &mockRevokeSessionTokenWhitelistRepo{}
	whitelistRepoMock.On("FindByUserID", mock.Anything, userID).Return([]domain.WhitelistEntry{
		{ID: tokenID, CreatedAt: now},
	}, nil)
	whitelistRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	cacheMock := &mockRevokeSessionTokenCache{}
	cacheMock.On("GetSessionToken", hash).Return(nil, false)
	cacheMock.On("DeleteSessionToken", hash).Return()

	config := AuthUsecaseConfig{
		TokenWhitelistSize: 10,
		ClockFunc:          func() time.Time { return now },
	}
	cmd := NewRevokeSessionTokenCommand(repoMock, whitelistRepoMock, cacheMock, config)

	input := &authservice.RevokeSessionTokenInput{RawToken: rawToken}

	// when
	err := cmd.RevokeSessionToken(context.Background(), input)

	// then
	assert.NoError(t, err)
	repoMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	cacheMock.AssertCalled(t, "DeleteSessionToken", hash)
	whitelistRepoMock.AssertCalled(t, "FindByUserID", mock.Anything, userID)
	whitelistRepoMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func Test_RevokeSessionTokenCommand_RevokeSessionToken_shouldReturnErrTokenRevoked_whenTokenIsAlreadyRevoked(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := now.Add(-5 * time.Minute)
	rawToken := "raw-session-token"
	hash := string(domain.HashToken(rawToken))
	tokenID := "session-token-id"

	sessionToken := domain.ReconstructSessionToken(tokenID, 1, "user1", "org1", domain.TokenHash(hash), now.Add(-1*time.Hour), now.Add(30*time.Minute), &revokedAt)

	repoMock := &mockRevokeSessionTokenRepo{}
	repoMock.On("FindByTokenHash", mock.Anything, hash).Return(sessionToken, nil)

	whitelistRepoMock := &mockRevokeSessionTokenWhitelistRepo{}

	cacheMock := &mockRevokeSessionTokenCache{}
	cacheMock.On("GetSessionToken", hash).Return(nil, false)

	config := AuthUsecaseConfig{TokenWhitelistSize: 10}
	cmd := NewRevokeSessionTokenCommand(repoMock, whitelistRepoMock, cacheMock, config)

	input := &authservice.RevokeSessionTokenInput{RawToken: rawToken}

	// when
	err := cmd.RevokeSessionToken(context.Background(), input)

	// then
	assert.ErrorIs(t, err, domain.ErrTokenRevoked)
	repoMock.AssertNotCalled(t, "Save")
}
