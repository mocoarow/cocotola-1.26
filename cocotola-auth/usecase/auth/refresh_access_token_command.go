package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type refreshAccessTokenRefreshRepo interface {
	FindByTokenHash(ctx context.Context, hash string) (*domaintoken.RefreshToken, error)
}

// RefreshAccessTokenCommand uses a raw refresh token to issue a new JWT access token.
type RefreshAccessTokenCommand struct {
	refreshRepo   refreshAccessTokenRefreshRepo
	accessRepo    accessTokenSaver
	whitelistRepo WhitelistRepository
	jwt           jwtCreator
	cache         accessTokenCacheSetter
	config        UsecaseConfig
}

// NewRefreshAccessTokenCommand returns a new RefreshAccessTokenCommand.
func NewRefreshAccessTokenCommand(
	refreshRepo refreshAccessTokenRefreshRepo,
	accessRepo accessTokenSaver,
	whitelistRepo WhitelistRepository,
	jwt jwtCreator,
	cache accessTokenCacheSetter,
	config UsecaseConfig,
) *RefreshAccessTokenCommand {
	return &RefreshAccessTokenCommand{
		refreshRepo:   refreshRepo,
		accessRepo:    accessRepo,
		whitelistRepo: whitelistRepo,
		jwt:           jwt,
		cache:         cache,
		config:        config,
	}
}

// RefreshAccessToken uses a raw refresh token to issue a new JWT access token.
func (c *RefreshAccessTokenCommand) RefreshAccessToken(ctx context.Context, input *authservice.RefreshAccessTokenInput) (*authservice.RefreshAccessTokenOutput, error) {
	hash := domaintoken.HashToken(input.RawRefreshToken)
	now := c.config.Now()

	refreshToken, err := c.refreshRepo.FindByTokenHash(ctx, string(hash))
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	if refreshToken.IsRevoked() {
		return nil, domain.ErrTokenRevoked
	}
	if refreshToken.IsExpired(now) {
		return nil, domain.ErrSessionExpired
	}

	accessExpiresAt := now.Add(time.Duration(c.config.AccessTokenTTLMin) * time.Minute)
	accessID := uuid.New().String()

	jwtString, err := c.jwt.CreateAccessToken(string(refreshToken.LoginID()), refreshToken.UserID(), refreshToken.OrganizationName(), accessID)
	if err != nil {
		return nil, fmt.Errorf("create jwt: %w", err)
	}

	accessToken, err := domaintoken.NewAccessToken(accessID, refreshToken.ID(), refreshToken.UserID(), refreshToken.LoginID(), refreshToken.OrganizationName(), now, accessExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("new access token: %w", err)
	}

	// TX1: Save access token
	if err := c.accessRepo.Save(ctx, accessToken); err != nil {
		return nil, fmt.Errorf("save access token: %w", err)
	}

	c.cache.SetAccessToken(accessID, accessToken)

	// TX2: Update access token whitelist
	entries, err := c.whitelistRepo.FindByUserID(ctx, refreshToken.UserID())
	if err != nil {
		return nil, fmt.Errorf("find access token whitelist entries: %w", err)
	}

	whitelist, err := domaintoken.NewWhitelist(refreshToken.UserID(), entries, c.config.TokenWhitelistSize)
	if err != nil {
		return nil, fmt.Errorf("new access token whitelist: %w", err)
	}
	whitelist.Add(domaintoken.WhitelistEntry{ID: accessID, CreatedAt: now})

	if err := c.whitelistRepo.Save(ctx, whitelist); err != nil {
		return nil, fmt.Errorf("save access token whitelist: %w", err)
	}

	output, err := authservice.NewRefreshAccessTokenOutput(jwtString)
	if err != nil {
		return nil, fmt.Errorf("create refresh access token output: %w", err)
	}
	return output, nil
}
