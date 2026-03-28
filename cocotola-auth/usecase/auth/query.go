package auth

import (
	"context"
	"time"

	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// SessionTokenRepository defines persistence operations for session tokens.
type SessionTokenRepository interface {
	Save(ctx context.Context, token *domaintoken.SessionToken) error
	FindByTokenHash(ctx context.Context, hash string) (*domaintoken.SessionToken, error)
}

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Save(ctx context.Context, token *domaintoken.RefreshToken) error
	FindByTokenHash(ctx context.Context, hash string) (*domaintoken.RefreshToken, error)
}

// AccessTokenRepository defines persistence operations for access tokens.
type AccessTokenRepository interface {
	Save(ctx context.Context, token *domaintoken.AccessToken) error
	FindByID(ctx context.Context, id string) (*domaintoken.AccessToken, error)
	FindByRefreshTokenID(ctx context.Context, refreshTokenID string) ([]domaintoken.AccessToken, error)
}

// WhitelistRepository defines persistence operations for token whitelists.
type WhitelistRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]domaintoken.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domaintoken.Whitelist) error
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
	SetSessionToken(hash string, token *domaintoken.SessionToken)
	GetSessionToken(hash string) (*domaintoken.SessionToken, bool)
	DeleteSessionToken(hash string)
	SetAccessToken(jti string, token *domaintoken.AccessToken)
	GetAccessToken(jti string) (*domaintoken.AccessToken, bool)
	DeleteAccessToken(jti string)
}

// UsecaseConfig holds TTL and whitelist configuration.
type UsecaseConfig struct {
	SessionTokenTTLMin int
	SessionMaxTTLMin   int
	AccessTokenTTLMin  int
	RefreshTokenTTLMin int
	TokenWhitelistSize int
	ClockFunc          func() time.Time
}

// Now returns the current time using ClockFunc if set, otherwise time.Now.
func (c UsecaseConfig) Now() time.Time {
	if c.ClockFunc != nil {
		return c.ClockFunc()
	}
	return time.Now()
}

// Query composes all authentication Query structs.
type Query struct {
	*PasswordAuthenticateQuery
	*GuestAuthenticateQuery
	*ValidateSessionTokenQuery
	*ValidateAccessTokenQuery
}

// NewQuery returns a new Query with the given dependencies.
func NewQuery(
	userAuthenticator UserAuthenticator,
	guestAuthenticator GuestAuthenticator,
	sessionTokenRepo SessionTokenRepository,
	sessionTokenWhitelistRepo WhitelistRepository,
	accessTokenRepo AccessTokenRepository,
	accessTokenWhitelistRepo WhitelistRepository,
	jwtManager JWTManager,
	tokenCache TokenCache,
	config UsecaseConfig,
) *Query {
	return &Query{
		PasswordAuthenticateQuery: NewPasswordAuthenticateQuery(userAuthenticator),
		GuestAuthenticateQuery:    NewGuestAuthenticateQuery(guestAuthenticator),
		ValidateSessionTokenQuery: NewValidateSessionTokenQuery(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		ValidateAccessTokenQuery:  NewValidateAccessTokenQuery(accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
	}
}
