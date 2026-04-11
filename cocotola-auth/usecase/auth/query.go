package auth

import (
	"context"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
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
	FindByUserID(ctx context.Context, userID domain.AppUserID) ([]domaintoken.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domaintoken.Whitelist) error
}

// JWTManager creates and parses JWT access tokens.
type JWTManager interface {
	CreateAccessToken(loginID string, userID domain.AppUserID, organizationName string, jti string) (string, error)
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

// SupabaseVerifier verifies Supabase JWT tokens and returns the user's sub
// (UUID) and email. Implementations MUST reject tokens whose email has not
// been verified by Supabase so the exchange flow can trust the returned email
// as a stable identifier for account linking.
type SupabaseVerifier interface {
	Verify(ctx context.Context, tokenString string) (sub string, email string, err error)
}

// AppUserProviderFinder finds an app user provider link by external provider ID.
type AppUserProviderFinder interface {
	FindByProviderID(ctx context.Context, organizationID domain.OrganizationID, provider string, providerID string) (*domainuser.AppUserProvider, error)
}

// AppUserProviderSaver persists an AppUserProvider entity.
type AppUserProviderSaver interface {
	Save(ctx context.Context, provider *domainuser.AppUserProvider) error
}

// AppUserByIDFinder finds an app user by ID.
type AppUserByIDFinder interface {
	FindByID(ctx context.Context, id domain.AppUserID) (*domainuser.AppUser, error)
}

// AppUserByLoginIDFinder finds an existing app user by organization and login ID.
type AppUserByLoginIDFinder interface {
	FindByLoginID(ctx context.Context, organizationID domain.OrganizationID, loginID domain.LoginID) (*domainuser.AppUser, error)
}

// AppUserSaver persists an app user aggregate as a whole.
type AppUserSaver interface {
	Save(ctx context.Context, user *domainuser.AppUser) error
}

// OrganizationFinder finds an organization by name.
type OrganizationFinder interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
}

// Query composes all authentication Query structs.
type Query struct {
	*PasswordAuthenticateQuery
	*GuestAuthenticateQuery
	*ValidateSessionTokenQuery
	*ValidateAccessTokenQuery
	*SupabaseExchangeQuery
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
	supabaseVerifier SupabaseVerifier,
	providerFinder AppUserProviderFinder,
	providerSaver AppUserProviderSaver,
	appUserByIDFinder AppUserByIDFinder,
	appUserByLoginIDFinder AppUserByLoginIDFinder,
	appUserSaver AppUserSaver,
	organizationFinder OrganizationFinder,
) *Query {
	return &Query{
		PasswordAuthenticateQuery: NewPasswordAuthenticateQuery(userAuthenticator),
		GuestAuthenticateQuery:    NewGuestAuthenticateQuery(guestAuthenticator),
		ValidateSessionTokenQuery: NewValidateSessionTokenQuery(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		ValidateAccessTokenQuery:  NewValidateAccessTokenQuery(accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		SupabaseExchangeQuery:     NewSupabaseExchangeQuery(supabaseVerifier, providerFinder, providerSaver, appUserByIDFinder, appUserByLoginIDFinder, appUserSaver, organizationFinder),
	}
}
