package auth

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type revokeTokenRefreshRepo interface {
	FindByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Save(ctx context.Context, token *domain.RefreshToken) error
}

type revokeTokenAccessRepo interface {
	FindByID(ctx context.Context, id string) (*domain.AccessToken, error)
	FindByRefreshTokenID(ctx context.Context, refreshTokenID string) ([]domain.AccessToken, error)
	Save(ctx context.Context, token *domain.AccessToken) error
}

type revokeTokenRefreshWhitelistRepo interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

type revokeTokenAccessWhitelistRepo interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

type revokeTokenJWT interface {
	ParseAccessToken(tokenString string) (*authservice.UserInfo, string, error)
}

type revokeTokenCache interface {
	DeleteAccessToken(jti string)
}

// RevokeTokenCommand revokes a token. If the token is a JWT, it revokes the access token.
// If the token is an opaque token, it revokes the refresh token and all associated access tokens.
type RevokeTokenCommand struct {
	refreshRepo          revokeTokenRefreshRepo
	accessRepo           revokeTokenAccessRepo
	refreshWhitelistRepo revokeTokenRefreshWhitelistRepo
	accessWhitelistRepo  revokeTokenAccessWhitelistRepo
	jwt                  revokeTokenJWT
	cache                revokeTokenCache
	config               AuthUsecaseConfig
}

// NewRevokeTokenCommand returns a new RevokeTokenCommand.
func NewRevokeTokenCommand(
	refreshRepo revokeTokenRefreshRepo,
	accessRepo revokeTokenAccessRepo,
	refreshWhitelistRepo revokeTokenRefreshWhitelistRepo,
	accessWhitelistRepo revokeTokenAccessWhitelistRepo,
	jwt revokeTokenJWT,
	cache revokeTokenCache,
	config AuthUsecaseConfig,
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

	hash := string(domain.HashToken(input.Token))
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
	entries, err := c.accessWhitelistRepo.FindByUserID(ctx, accessToken.UserID())
	if err != nil {
		return fmt.Errorf("find access token whitelist: %w", err)
	}
	whitelist := domain.NewTokenWhitelist(accessToken.UserID(), entries, c.config.TokenWhitelistSize)
	whitelist.Remove([]string{jti})
	if err := c.accessWhitelistRepo.Save(ctx, whitelist); err != nil {
		return fmt.Errorf("save access token whitelist: %w", err)
	}
	return nil
}

func (c *RevokeTokenCommand) revokeRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error {
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
	refreshEntries, err := c.refreshWhitelistRepo.FindByUserID(ctx, refreshToken.UserID())
	if err != nil {
		return fmt.Errorf("find refresh token whitelist: %w", err)
	}
	refreshWhitelist := domain.NewTokenWhitelist(refreshToken.UserID(), refreshEntries, c.config.TokenWhitelistSize)
	refreshWhitelist.Remove([]string{refreshToken.ID()})
	if err := c.refreshWhitelistRepo.Save(ctx, refreshWhitelist); err != nil {
		return fmt.Errorf("save refresh token whitelist: %w", err)
	}

	// TX6: Remove access tokens from whitelist
	if len(revokedAccessIDs) > 0 {
		accessEntries, err := c.accessWhitelistRepo.FindByUserID(ctx, refreshToken.UserID())
		if err != nil {
			return fmt.Errorf("find access token whitelist: %w", err)
		}
		accessWhitelist := domain.NewTokenWhitelist(refreshToken.UserID(), accessEntries, c.config.TokenWhitelistSize)
		accessWhitelist.Remove(revokedAccessIDs)
		if err := c.accessWhitelistRepo.Save(ctx, accessWhitelist); err != nil {
			return fmt.Errorf("save access token whitelist: %w", err)
		}
	}

	return nil
}
