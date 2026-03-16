package auth

import (
	"context"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// SessionTokenRepository defines persistence operations for session tokens.
type SessionTokenRepository interface {
	Save(ctx context.Context, token *domain.SessionToken) error
	FindByTokenHash(ctx context.Context, hash string) (*domain.SessionToken, error)
}

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	FindByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
}

// AccessTokenRepository defines persistence operations for access tokens.
type AccessTokenRepository interface {
	Save(ctx context.Context, token *domain.AccessToken) error
	FindByID(ctx context.Context, id string) (*domain.AccessToken, error)
	FindByRefreshTokenID(ctx context.Context, refreshTokenID string) ([]domain.AccessToken, error)
}

// SessionTokenWhitelistRepository defines persistence operations for session token whitelist.
type SessionTokenWhitelistRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

// RefreshTokenWhitelistRepository defines persistence operations for refresh token whitelist.
type RefreshTokenWhitelistRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

// AccessTokenWhitelistRepository defines persistence operations for access token whitelist.
type AccessTokenWhitelistRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

// JWTManager creates and parses JWT access tokens.
type JWTManager interface {
	CreateAccessToken(loginID string, userID int, organizationName string, jti string) (string, error)
	ParseAccessToken(tokenString string) (*authservice.UserInfo, string, error)
}

// UserAuthenticator verifies user credentials and returns user info.
type UserAuthenticator interface {
	Authenticate(ctx context.Context, loginID string, password string, organizationName string) (*authservice.UserInfo, error)
}

// TokenCache provides in-memory caching for session and access tokens.
type TokenCache interface {
	SetSessionToken(hash string, token *domain.SessionToken)
	GetSessionToken(hash string) (*domain.SessionToken, bool)
	DeleteSessionToken(hash string)
	SetAccessToken(jti string, token *domain.AccessToken)
	GetAccessToken(jti string) (*domain.AccessToken, bool)
	DeleteAccessToken(jti string)
}

// AuthUsecaseConfig holds TTL and whitelist configuration.
type AuthUsecaseConfig struct {
	SessionTokenTTLMin int
	SessionMaxTTLMin   int
	AccessTokenTTLMin  int
	RefreshTokenTTLMin int
	TokenWhitelistSize int
	ClockFunc          func() time.Time
}

// Now returns the current time using ClockFunc if set, otherwise time.Now.
func (c AuthUsecaseConfig) Now() time.Time {
	if c.ClockFunc != nil {
		return c.ClockFunc()
	}
	return time.Now()
}

// AuthQuery composes all authentication Query structs.
type AuthQuery struct {
	*PasswordAuthenticateQuery
	*ValidateSessionTokenQuery
	*ValidateAccessTokenQuery
}

// NewAuthQuery returns a new AuthQuery with the given dependencies.
func NewAuthQuery(
	userAuthenticator UserAuthenticator,
	sessionTokenRepo SessionTokenRepository,
	sessionTokenWhitelistRepo SessionTokenWhitelistRepository,
	accessTokenRepo AccessTokenRepository,
	accessTokenWhitelistRepo AccessTokenWhitelistRepository,
	jwtManager JWTManager,
	tokenCache TokenCache,
	config AuthUsecaseConfig,
) *AuthQuery {
	return &AuthQuery{
		PasswordAuthenticateQuery: NewPasswordAuthenticateQuery(userAuthenticator),
		ValidateSessionTokenQuery: NewValidateSessionTokenQuery(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		ValidateAccessTokenQuery:  NewValidateAccessTokenQuery(accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
	}
}
