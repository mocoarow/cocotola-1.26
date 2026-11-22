package auth

import "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

// Command composes all authentication Command structs.
type Command struct {
	*CreateSessionTokenCommand
	*CreateTokenPairCommand
	*ExtendSessionTokenCommand
	*RevokeSessionTokenCommand
	*RefreshAccessTokenCommand
	*RevokeTokenCommand
	*SupabaseExchangeCommand
}

// NewCommand returns a new Command with the given dependencies.
func NewCommand(
	sessionTokenRepo SessionTokenRepository,
	sessionTokenWhitelistRepo WhitelistRepository,
	refreshTokenRepo RefreshTokenRepository,
	refreshTokenWhitelistRepo WhitelistRepository,
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
	publisher domain.EventPublisher,
) *Command {
	return &Command{
		CreateSessionTokenCommand: NewCreateSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		CreateTokenPairCommand:    NewCreateTokenPairCommand(refreshTokenRepo, accessTokenRepo, refreshTokenWhitelistRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		ExtendSessionTokenCommand: NewExtendSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		RevokeSessionTokenCommand: NewRevokeSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		RefreshAccessTokenCommand: NewRefreshAccessTokenCommand(refreshTokenRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		RevokeTokenCommand:        NewRevokeTokenCommand(refreshTokenRepo, accessTokenRepo, refreshTokenWhitelistRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		SupabaseExchangeCommand:   NewSupabaseExchangeCommand(supabaseVerifier, providerFinder, providerSaver, appUserByIDFinder, appUserByLoginIDFinder, appUserSaver, organizationFinder, publisher),
	}
}

// Usecase composes all authentication Command and Query structs.
// Each embedded struct provides a single focused operation with only
// the dependencies it needs.
type Usecase struct {
	*Query
	*Command
}

// NewUsecase returns a new Usecase with the given dependencies.
func NewUsecase(
	groupFinder GroupFinder,
	passwordHasher PasswordHasher,
	sessionTokenRepo SessionTokenRepository,
	sessionTokenWhitelistRepo WhitelistRepository,
	refreshTokenRepo RefreshTokenRepository,
	refreshTokenWhitelistRepo WhitelistRepository,
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
	publisher domain.EventPublisher,
) *Usecase {
	return &Usecase{
		Query:   NewQuery(organizationFinder, appUserByLoginIDFinder, groupFinder, passwordHasher, sessionTokenRepo, sessionTokenWhitelistRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		Command: NewCommand(sessionTokenRepo, sessionTokenWhitelistRepo, refreshTokenRepo, refreshTokenWhitelistRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config, supabaseVerifier, providerFinder, providerSaver, appUserByIDFinder, appUserByLoginIDFinder, appUserSaver, organizationFinder, publisher),
	}
}
