package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type createTokenPairRefreshRepo interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
}

type createTokenPairAccessRepo interface {
	Save(ctx context.Context, token *domain.AccessToken) error
}

type createTokenPairRefreshWhitelistRepo interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

type createTokenPairAccessWhitelistRepo interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
	Save(ctx context.Context, whitelist *domain.TokenWhitelist) error
}

type createTokenPairJWT interface {
	CreateAccessToken(loginID string, userID int, organizationName string, jti string) (string, error)
}

type createTokenPairCache interface {
	SetAccessToken(jti string, token *domain.AccessToken)
}

// CreateTokenPairCommand creates a new access token (JWT) and refresh token (opaque) pair.
type CreateTokenPairCommand struct {
	refreshRepo          createTokenPairRefreshRepo
	accessRepo           createTokenPairAccessRepo
	refreshWhitelistRepo createTokenPairRefreshWhitelistRepo
	accessWhitelistRepo  createTokenPairAccessWhitelistRepo
	jwt                  createTokenPairJWT
	cache                createTokenPairCache
	config               AuthUsecaseConfig
}

// NewCreateTokenPairCommand returns a new CreateTokenPairCommand.
func NewCreateTokenPairCommand(
	refreshRepo createTokenPairRefreshRepo,
	accessRepo createTokenPairAccessRepo,
	refreshWhitelistRepo createTokenPairRefreshWhitelistRepo,
	accessWhitelistRepo createTokenPairAccessWhitelistRepo,
	jwt createTokenPairJWT,
	cache createTokenPairCache,
	config AuthUsecaseConfig,
) *CreateTokenPairCommand {
	return &CreateTokenPairCommand{
		refreshRepo:          refreshRepo,
		accessRepo:           accessRepo,
		refreshWhitelistRepo: refreshWhitelistRepo,
		accessWhitelistRepo:  accessWhitelistRepo,
		jwt:                  jwt,
		cache:                cache,
		config:               config,
	}
}

// CreateTokenPair creates a new access token (JWT) and refresh token (opaque) pair.
func (c *CreateTokenPairCommand) CreateTokenPair(ctx context.Context, input *authservice.CreateTokenPairInput) (*authservice.CreateTokenPairOutput, error) {
	rawRefresh, refreshHash, err := domain.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	now := c.config.Now()
	refreshExpiresAt := now.Add(time.Duration(c.config.RefreshTokenTTLMin) * time.Minute)
	refreshID := uuid.New().String()

	refreshToken, err := domain.NewRefreshToken(refreshID, input.UserID, domain.LoginID(input.LoginID), input.OrganizationName, refreshHash, now, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("new refresh token: %w", err)
	}

	accessExpiresAt := now.Add(time.Duration(c.config.AccessTokenTTLMin) * time.Minute)
	accessID := uuid.New().String()

	jwtString, err := c.jwt.CreateAccessToken(input.LoginID, input.UserID, input.OrganizationName, accessID)
	if err != nil {
		return nil, fmt.Errorf("create jwt: %w", err)
	}

	accessToken, err := domain.NewAccessToken(accessID, refreshID, input.UserID, domain.LoginID(input.LoginID), input.OrganizationName, now, accessExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("new access token: %w", err)
	}

	// TX1: Save refresh token
	if err := c.refreshRepo.Save(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	// TX2: Save access token
	if err := c.accessRepo.Save(ctx, accessToken); err != nil {
		return nil, fmt.Errorf("save access token: %w", err)
	}

	c.cache.SetAccessToken(accessID, accessToken)

	// TX3: Update refresh token whitelist
	refreshEntries, err := c.refreshWhitelistRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("find refresh token whitelist entries: %w", err)
	}

	refreshWhitelist := domain.NewTokenWhitelist(input.UserID, refreshEntries, c.config.TokenWhitelistSize)
	refreshWhitelist.Add(domain.WhitelistEntry{ID: refreshID, CreatedAt: now})

	if err := c.refreshWhitelistRepo.Save(ctx, refreshWhitelist); err != nil {
		return nil, fmt.Errorf("save refresh token whitelist: %w", err)
	}

	// TX4: Update access token whitelist
	accessEntries, err := c.accessWhitelistRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("find access token whitelist entries: %w", err)
	}

	accessWhitelist := domain.NewTokenWhitelist(input.UserID, accessEntries, c.config.TokenWhitelistSize)
	accessWhitelist.Add(domain.WhitelistEntry{ID: accessID, CreatedAt: now})

	if err := c.accessWhitelistRepo.Save(ctx, accessWhitelist); err != nil {
		return nil, fmt.Errorf("save access token whitelist: %w", err)
	}

	output, err := authservice.NewCreateTokenPairOutput(jwtString, rawRefresh)
	if err != nil {
		return nil, fmt.Errorf("create token pair output: %w", err)
	}
	return output, nil
}
