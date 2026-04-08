package auth

// Command composes all authentication Command structs.
type Command struct {
	*CreateSessionTokenCommand
	*CreateTokenPairCommand
	*ExtendSessionTokenCommand
	*RevokeSessionTokenCommand
	*RefreshAccessTokenCommand
	*RevokeTokenCommand
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
) *Command {
	return &Command{
		CreateSessionTokenCommand: NewCreateSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		CreateTokenPairCommand:    NewCreateTokenPairCommand(refreshTokenRepo, accessTokenRepo, refreshTokenWhitelistRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		ExtendSessionTokenCommand: NewExtendSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		RevokeSessionTokenCommand: NewRevokeSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		RefreshAccessTokenCommand: NewRefreshAccessTokenCommand(refreshTokenRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		RevokeTokenCommand:        NewRevokeTokenCommand(refreshTokenRepo, accessTokenRepo, refreshTokenWhitelistRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
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
	userAuthenticator UserAuthenticator,
	guestAuthenticator GuestAuthenticator,
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
	appUserFinder AppUserProviderFinder,
	appUserIDProvider AppUserIDProvider,
	appUserByLoginIDFinder AppUserByLoginIDFinder,
	appUserSaver AppUserSaver,
	organizationFinder OrganizationFinder,
) *Usecase {
	return &Usecase{
		Query:   NewQuery(userAuthenticator, guestAuthenticator, sessionTokenRepo, sessionTokenWhitelistRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config, supabaseVerifier, appUserFinder, appUserIDProvider, appUserByLoginIDFinder, appUserSaver, organizationFinder),
		Command: NewCommand(sessionTokenRepo, sessionTokenWhitelistRepo, refreshTokenRepo, refreshTokenWhitelistRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
	}
}
