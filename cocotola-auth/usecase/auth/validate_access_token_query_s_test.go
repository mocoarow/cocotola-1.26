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

type mockValidateAccessTokenRepo struct{ mock.Mock }

func (m *mockValidateAccessTokenRepo) FindByID(ctx context.Context, id string) (*domain.AccessToken, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AccessToken), args.Error(1)
}

type mockValidateAccessTokenWhitelistRepo struct{ mock.Mock }

func (m *mockValidateAccessTokenWhitelistRepo) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.WhitelistEntry), args.Error(1)
}

type mockValidateAccessTokenJWT struct{ mock.Mock }

func (m *mockValidateAccessTokenJWT) ParseAccessToken(tokenString string) (*authservice.UserInfo, string, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*authservice.UserInfo), args.String(1), args.Error(2)
}

type mockValidateAccessTokenCache struct {
	mock.Mock
}

func (m *mockValidateAccessTokenCache) GetAccessToken(jti string) (*domain.AccessToken, bool) {
	args := m.Called(jti)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*domain.AccessToken), args.Bool(1)
}

func (m *mockValidateAccessTokenCache) SetAccessToken(jti string, token *domain.AccessToken) {
	m.Called(jti, token)
}

func (m *mockValidateAccessTokenCache) DeleteAccessToken(jti string) {
	m.Called(jti)
}

// --- tests ---

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnUserInfo_whenTokenIsValidInCache(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	jti := "access-token-id"

	userInfo, _ := authservice.NewUserInfo(1, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domain.ReconstructAccessToken(jti, "refresh-1", 1, "user1", "org1", now.Add(-30*time.Minute), now.Add(30*time.Minute), nil)

	jwtMock := &mockValidateAccessTokenJWT{}
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := &mockValidateAccessTokenCache{}
	cacheMock.On("GetAccessToken", jti).Return(accessToken, true)

	repoMock := &mockValidateAccessTokenRepo{}
	whitelistRepoMock := &mockValidateAccessTokenWhitelistRepo{}

	query := NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, AuthUsecaseConfig{
		ClockFunc: func() time.Time { return now },
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	output, err := query.ValidateAccessToken(context.Background(), input)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1, output.UserID)
	assert.Equal(t, "user1", output.LoginID)
	assert.Equal(t, "org1", output.OrganizationName)
	repoMock.AssertNotCalled(t, "FindByID")
}

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnUserInfo_whenTokenIsValidInDB(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	jti := "access-token-id"
	userID := 1

	userInfo, _ := authservice.NewUserInfo(userID, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domain.ReconstructAccessToken(jti, "refresh-1", userID, "user1", "org1", now.Add(-30*time.Minute), now.Add(30*time.Minute), nil)

	jwtMock := &mockValidateAccessTokenJWT{}
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := &mockValidateAccessTokenCache{}
	cacheMock.On("GetAccessToken", jti).Return(nil, false)
	cacheMock.On("SetAccessToken", jti, accessToken).Return()

	repoMock := &mockValidateAccessTokenRepo{}
	repoMock.On("FindByID", mock.Anything, jti).Return(accessToken, nil)

	whitelistRepoMock := &mockValidateAccessTokenWhitelistRepo{}
	whitelistRepoMock.On("FindByUserID", mock.Anything, userID).Return([]domain.WhitelistEntry{
		{ID: jti, CreatedAt: now},
	}, nil)

	query := NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, AuthUsecaseConfig{
		ClockFunc:          func() time.Time { return now },
		TokenWhitelistSize: 10,
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	output, err := query.ValidateAccessToken(context.Background(), input)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1, output.UserID)
	cacheMock.AssertCalled(t, "SetAccessToken", jti, accessToken)
}

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnErrTokenRevoked_whenTokenIsRevokedInCache(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := now.Add(-10 * time.Minute)
	jti := "access-token-id"

	userInfo, _ := authservice.NewUserInfo(1, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domain.ReconstructAccessToken(jti, "refresh-1", 1, "user1", "org1", now.Add(-30*time.Minute), now.Add(30*time.Minute), &revokedAt)

	jwtMock := &mockValidateAccessTokenJWT{}
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := &mockValidateAccessTokenCache{}
	cacheMock.On("GetAccessToken", jti).Return(accessToken, true)
	cacheMock.On("DeleteAccessToken", jti).Return()

	repoMock := &mockValidateAccessTokenRepo{}
	whitelistRepoMock := &mockValidateAccessTokenWhitelistRepo{}

	query := NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, AuthUsecaseConfig{
		ClockFunc: func() time.Time { return now },
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	_, err := query.ValidateAccessToken(context.Background(), input)

	// then
	assert.ErrorIs(t, err, domain.ErrTokenRevoked)
	cacheMock.AssertCalled(t, "DeleteAccessToken", jti)
}

func Test_ValidateAccessTokenQuery_ValidateAccessToken_shouldReturnErrSessionExpired_whenTokenIsExpired(t *testing.T) {
	t.Parallel()

	// given
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	jti := "access-token-id"

	userInfo, _ := authservice.NewUserInfo(1, "user1", "org1", now.Add(1*time.Hour))
	accessToken := domain.ReconstructAccessToken(jti, "refresh-1", 1, "user1", "org1", now.Add(-2*time.Hour), now.Add(-1*time.Hour), nil)

	jwtMock := &mockValidateAccessTokenJWT{}
	jwtMock.On("ParseAccessToken", "jwt-string").Return(userInfo, jti, nil)

	cacheMock := &mockValidateAccessTokenCache{}
	cacheMock.On("GetAccessToken", jti).Return(accessToken, true)

	repoMock := &mockValidateAccessTokenRepo{}
	whitelistRepoMock := &mockValidateAccessTokenWhitelistRepo{}

	query := NewValidateAccessTokenQuery(repoMock, whitelistRepoMock, jwtMock, cacheMock, AuthUsecaseConfig{
		ClockFunc: func() time.Time { return now },
	})

	input := &authservice.ValidateAccessTokenInput{JWTString: "jwt-string"}

	// when
	_, err := query.ValidateAccessToken(context.Background(), input)

	// then
	assert.ErrorIs(t, err, domain.ErrSessionExpired)
}
