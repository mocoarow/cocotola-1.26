package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

func Test_RevokeSessionTokenCommand_RevokeSessionToken_shouldRevokeToken_whenTokenIsActive(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	rawInput := "raw-session-value"
	hash := string(domaintoken.HashToken(rawInput))
	tokenID := "session-token-id"
	userID := fixtureAppUserID

	sessionToken := domaintoken.ReconstructSessionToken(tokenID, userID, "user1", "org1", domain.TokenHash(hash), now, now.Add(30*time.Minute), nil)

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, hash).Return(sessionToken, nil)
	repoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	whitelistRepoMock := NewMockWhitelistRepository(t)
	whitelistRepoMock.On("FindByUserID", mock.Anything, userID).Return([]domaintoken.WhitelistEntry{
		{ID: tokenID, CreatedAt: now},
	}, nil)
	whitelistRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", hash).Return(nil, false)
	cacheMock.On("DeleteSessionToken", hash).Return()

	config := authusecase.UsecaseConfig{
		TokenWhitelistSize: 10,
		ClockFunc:          func() time.Time { return now },
	}
	cmd := authusecase.NewRevokeSessionTokenCommand(repoMock, whitelistRepoMock, cacheMock, config)

	input := &authservice.RevokeSessionTokenInput{RawToken: rawInput}

	// when
	err := cmd.RevokeSessionToken(context.Background(), input)

	// then
	require.NoError(t, err)
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
	rawInput := "raw-session-value"
	hash := string(domaintoken.HashToken(rawInput))
	tokenID := "session-token-id"

	sessionToken := domaintoken.ReconstructSessionToken(tokenID, fixtureAppUserID, "user1", "org1", domain.TokenHash(hash), now.Add(-1*time.Hour), now.Add(30*time.Minute), &revokedAt)

	repoMock := NewMockSessionTokenRepository(t)
	repoMock.On("FindByTokenHash", mock.Anything, hash).Return(sessionToken, nil)

	whitelistRepoMock := NewMockWhitelistRepository(t)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetSessionToken", hash).Return(nil, false)

	config := authusecase.UsecaseConfig{TokenWhitelistSize: 10}
	cmd := authusecase.NewRevokeSessionTokenCommand(repoMock, whitelistRepoMock, cacheMock, config)

	input := &authservice.RevokeSessionTokenInput{RawToken: rawInput}

	// when
	err := cmd.RevokeSessionToken(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrTokenRevoked)
	repoMock.AssertNotCalled(t, "Save")
}
