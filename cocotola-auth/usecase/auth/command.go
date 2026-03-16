package auth

// AuthCommand composes all authentication Command structs.
type AuthCommand struct {
	*CreateSessionTokenCommand
	*CreateTokenPairCommand
	*ExtendSessionTokenCommand
	*RevokeSessionTokenCommand
	*RefreshAccessTokenCommand
	*RevokeTokenCommand
}

// NewAuthCommand returns a new AuthCommand with the given dependencies.
func NewAuthCommand(
	sessionTokenRepo SessionTokenRepository,
	sessionTokenWhitelistRepo SessionTokenWhitelistRepository,
	refreshTokenRepo RefreshTokenRepository,
	refreshTokenWhitelistRepo RefreshTokenWhitelistRepository,
	accessTokenRepo AccessTokenRepository,
	accessTokenWhitelistRepo AccessTokenWhitelistRepository,
	jwtManager JWTManager,
	tokenCache TokenCache,
	config AuthUsecaseConfig,
) *AuthCommand {
	return &AuthCommand{
		CreateSessionTokenCommand: NewCreateSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		CreateTokenPairCommand:    NewCreateTokenPairCommand(refreshTokenRepo, accessTokenRepo, refreshTokenWhitelistRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		ExtendSessionTokenCommand: NewExtendSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		RevokeSessionTokenCommand: NewRevokeSessionTokenCommand(sessionTokenRepo, sessionTokenWhitelistRepo, tokenCache, config),
		RefreshAccessTokenCommand: NewRefreshAccessTokenCommand(refreshTokenRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		RevokeTokenCommand:        NewRevokeTokenCommand(refreshTokenRepo, accessTokenRepo, refreshTokenWhitelistRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
	}
}

// AuthUsecase composes all authentication Command and Query structs.
// Each embedded struct provides a single focused operation with only
// the dependencies it needs.
type AuthUsecase struct {
	*AuthQuery
	*AuthCommand
}

// NewAuthUsecase returns a new AuthUsecase with the given dependencies.
func NewAuthUsecase(
	userAuthenticator UserAuthenticator,
	sessionTokenRepo SessionTokenRepository,
	sessionTokenWhitelistRepo SessionTokenWhitelistRepository,
	refreshTokenRepo RefreshTokenRepository,
	refreshTokenWhitelistRepo RefreshTokenWhitelistRepository,
	accessTokenRepo AccessTokenRepository,
	accessTokenWhitelistRepo AccessTokenWhitelistRepository,
	jwtManager JWTManager,
	tokenCache TokenCache,
	config AuthUsecaseConfig,
) *AuthUsecase {
	return &AuthUsecase{
		AuthQuery:   NewAuthQuery(userAuthenticator, sessionTokenRepo, sessionTokenWhitelistRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
		AuthCommand: NewAuthCommand(sessionTokenRepo, sessionTokenWhitelistRepo, refreshTokenRepo, refreshTokenWhitelistRepo, accessTokenRepo, accessTokenWhitelistRepo, jwtManager, tokenCache, config),
	}
}
