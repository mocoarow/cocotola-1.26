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
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnUserInfo_whenTokenIsValidInCache(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	jti := "access-token-id"
	userID := fixtureAppUserID

	userInfo, _ := authservice.NewUserInfo(userID, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domaintoken.ReconstructAccessToken(jti, "refresh-1", userID, "user1", "org1", now.Add(-30*time.Minute), now.Add(30*time.Minute), nil)

	jwtMock := NewMockJWTManager(t)
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetAccessToken", jti).Return(accessToken, true)

	repoMock := NewMockAccessTokenRepository(t)
	whitelistRepoMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, authusecase.UsecaseConfig{
		ClockFunc: func() time.Time { return now },
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	output, err := query.ValidateAccessToken(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.True(t, userID.Equal(output.UserID))
	assert.Equal(t, "user1", output.LoginID)
	assert.Equal(t, "org1", output.OrganizationName)
	repoMock.AssertNotCalled(t, "FindByID")
}

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnUserInfo_whenTokenIsValidInDB(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	jti := "access-token-id"
	userID := fixtureAppUserID

	userInfo, _ := authservice.NewUserInfo(userID, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domaintoken.ReconstructAccessToken(jti, "refresh-1", userID, "user1", "org1", now.Add(-30*time.Minute), now.Add(30*time.Minute), nil)

	jwtMock := NewMockJWTManager(t)
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetAccessToken", jti).Return(nil, false)
	cacheMock.On("SetAccessToken", jti, accessToken).Return()

	repoMock := NewMockAccessTokenRepository(t)
	repoMock.On("FindByID", mock.Anything, jti).Return(accessToken, nil)

	whitelistRepoMock := NewMockWhitelistRepository(t)
	whitelistRepoMock.On("FindByUserID", mock.Anything, userID).Return([]domaintoken.WhitelistEntry{
		{ID: jti, CreatedAt: now},
	}, nil)

	query := authusecase.NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, authusecase.UsecaseConfig{
		ClockFunc:          func() time.Time { return now },
		TokenWhitelistSize: 10,
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	output, err := query.ValidateAccessToken(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.True(t, userID.Equal(output.UserID))
	cacheMock.AssertCalled(t, "SetAccessToken", jti, accessToken)
}

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnErrTokenRevoked_whenTokenIsRevokedInCache(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := now.Add(-10 * time.Minute)
	jti := "access-token-id"
	userID := fixtureAppUserID

	userInfo, _ := authservice.NewUserInfo(userID, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domaintoken.ReconstructAccessToken(jti, "refresh-1", userID, "user1", "org1", now.Add(-30*time.Minute), now.Add(30*time.Minute), &revokedAt)

	jwtMock := NewMockJWTManager(t)
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetAccessToken", jti).Return(accessToken, true)
	cacheMock.On("DeleteAccessToken", jti).Return()

	repoMock := NewMockAccessTokenRepository(t)
	whitelistRepoMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, authusecase.UsecaseConfig{
		ClockFunc: func() time.Time { return now },
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	_, err := query.ValidateAccessToken(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrTokenRevoked)
	cacheMock.AssertCalled(t, "DeleteAccessToken", jti)
}

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnErrSessionExpired_whenTokenIsExpired(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	jti := "access-token-id"
	userID := fixtureAppUserID

	userInfo, _ := authservice.NewUserInfo(userID, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domaintoken.ReconstructAccessToken(jti, "refresh-1", userID, "user1", "org1", now.Add(-2*time.Hour), now.Add(-1*time.Hour), nil)

	jwtMock := NewMockJWTManager(t)
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := NewMockTokenCache(t)
	cacheMock.On("GetAccessToken", jti).Return(accessToken, true)

	repoMock := NewMockAccessTokenRepository(t)
	whitelistRepoMock := NewMockWhitelistRepository(t)

	query := authusecase.NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, authusecase.UsecaseConfig{
		ClockFunc: func() time.Time { return now },
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	_, err := query.ValidateAccessToken(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrSessionExpired)
}
