package auth

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type revokeTokenCache interface {
	DeleteAccessToken(jti string)
}

// RevokeTokenCommand revokes a token. If the token is a JWT, it revokes the access token.
// If the token is an opaque token, it revokes the refresh token and all associated access tokens.
type RevokeTokenCommand struct {
	refreshRepo          RefreshTokenRepository
	accessRepo           AccessTokenRepository
	refreshWhitelistRepo WhitelistRepository
	accessWhitelistRepo  WhitelistRepository
	jwt                  jwtParser
	cache                revokeTokenCache
	config               UsecaseConfig
}

// NewRevokeTokenCommand returns a new RevokeTokenCommand.
func NewRevokeTokenCommand(
	refreshRepo RefreshTokenRepository,
	accessRepo AccessTokenRepository,
	refreshWhitelistRepo WhitelistRepository,
	accessWhitelistRepo WhitelistRepository,
	jwt jwtParser,
	cache revokeTokenCache,
	config UsecaseConfig,
) *RevokeTokenCommand {
	return &RevokeTokenCommand{
		refreshRepo:          refreshRepo,
		accessRepo:           accessRepo,
		refreshWhitelistRepo: refreshWhitelistRepo,
		accessWhitelistRepo:  accessWhitelistRepo,
		jwt:                  jwt,
		cache:                cache,
		config:               config,
	}
}

// RevokeToken revokes a token. If the token is a JWT, it revokes the access token.
// If the token is an opaque token, it revokes the refresh token and all associated access tokens.
func (c *RevokeTokenCommand) RevokeToken(ctx context.Context, input *authservice.RevokeTokenInput) error {
	_, jti, err := c.jwt.ParseAccessToken(input.Token)
	if err == nil {
		return c.revokeAccessToken(ctx, jti)
	}

	hash := string(domaintoken.HashToken(input.Token))
	refreshToken, err := c.refreshRepo.FindByTokenHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("find refresh token: %w", err)
	}

	return c.revokeRefreshToken(ctx, refreshToken)
}

func (c *RevokeTokenCommand) revokeAccessToken(ctx context.Context, jti string) error {
	// TX1: Load, revoke, and save access token
	accessToken, err := c.accessRepo.FindByID(ctx, jti)
	if err != nil {
		return fmt.Errorf("find access token: %w", err)
	}

	now := c.config.Now()
	accessToken.Revoke(now)
	if err := c.accessRepo.Save(ctx, accessToken); err != nil {
		return fmt.Errorf("save access token: %w", err)
	}
	c.cache.DeleteAccessToken(jti)

	// TX2: Remove from access token whitelist
	if err := c.removeFromWhitelist(ctx, c.accessWhitelistRepo, accessToken.UserID(), []string{jti}); err != nil {
		return fmt.Errorf("remove access token from whitelist: %w", err)
	}
	return nil
}

func (c *RevokeTokenCommand) revokeRefreshToken(ctx context.Context, refreshToken *domaintoken.RefreshToken) error {
	accessTokens, err := c.accessRepo.FindByRefreshTokenID(ctx, refreshToken.ID())
	if err != nil {
		return fmt.Errorf("find access tokens by refresh token id: %w", err)
	}

	// TX3: Revoke refresh token
	now := c.config.Now()
	refreshToken.Revoke(now)
	if err := c.refreshRepo.Save(ctx, refreshToken); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	// TX4: Revoke each access token individually
	var revokedAccessIDs []string
	for i := range accessTokens {
		if accessTokens[i].IsRevoked() {
			continue
		}
		accessTokens[i].Revoke(now)
		if err := c.accessRepo.Save(ctx, &accessTokens[i]); err != nil {
			return fmt.Errorf("save access token %s: %w", accessTokens[i].ID(), err)
		}
		c.cache.DeleteAccessToken(accessTokens[i].ID())
		revokedAccessIDs = append(revokedAccessIDs, accessTokens[i].ID())
	}

	// TX5: Remove refresh token from whitelist
	if err := c.removeFromWhitelist(ctx, c.refreshWhitelistRepo, refreshToken.UserID(), []string{refreshToken.ID()}); err != nil {
		return fmt.Errorf("remove refresh token from whitelist: %w", err)
	}

	// TX6: Remove access tokens from whitelist
	if len(revokedAccessIDs) > 0 {
		if err := c.removeFromWhitelist(ctx, c.accessWhitelistRepo, refreshToken.UserID(), revokedAccessIDs); err != nil {
			return fmt.Errorf("remove access tokens from whitelist: %w", err)
		}
	}

	return nil
}

func (c *RevokeTokenCommand) removeFromWhitelist(ctx context.Context, repo WhitelistRepository, userID domain.AppUserID, ids []string) error {
	entries, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find whitelist entries: %w", err)
	}
	whitelist, err := domaintoken.NewWhitelist(userID, entries, c.config.TokenWhitelistSize)
	if err != nil {
		return fmt.Errorf("new whitelist: %w", err)
	}
	whitelist.Remove(ids)
	if err := repo.Save(ctx, whitelist); err != nil {
		return fmt.Errorf("save whitelist: %w", err)
	}
	return nil
}
