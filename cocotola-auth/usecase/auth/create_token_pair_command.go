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

type createTokenPairRefreshRepo interface {
	Save(ctx context.Context, token *domaintoken.RefreshToken) error
}

// CreateTokenPairCommand creates a new access token (JWT) and refresh token (opaque) pair.
type CreateTokenPairCommand struct {
	refreshRepo          createTokenPairRefreshRepo
	accessRepo           accessTokenSaver
	refreshWhitelistRepo WhitelistRepository
	accessWhitelistRepo  WhitelistRepository
	jwt                  jwtCreator
	cache                accessTokenCacheSetter
	config               UsecaseConfig
}

// NewCreateTokenPairCommand returns a new CreateTokenPairCommand.
func NewCreateTokenPairCommand(
	refreshRepo createTokenPairRefreshRepo,
	accessRepo accessTokenSaver,
	refreshWhitelistRepo WhitelistRepository,
	accessWhitelistRepo WhitelistRepository,
	jwt jwtCreator,
	cache accessTokenCacheSetter,
	config UsecaseConfig,
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
	rawRefresh, refreshHash, err := domaintoken.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	now := c.config.Now()
	refreshExpiresAt := now.Add(time.Duration(c.config.RefreshTokenTTLMin) * time.Minute)
	refreshID := uuid.New().String()

	refreshToken, err := domaintoken.NewRefreshToken(refreshID, input.UserID, domain.LoginID(input.LoginID), input.OrganizationName, refreshHash, now, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("new refresh token: %w", err)
	}

	accessExpiresAt := now.Add(time.Duration(c.config.AccessTokenTTLMin) * time.Minute)
	accessID := uuid.New().String()

	jwtString, err := c.jwt.CreateAccessToken(input.LoginID, input.UserID, input.OrganizationName, accessID)
	if err != nil {
		return nil, fmt.Errorf("create jwt: %w", err)
	}

	accessToken, err := domaintoken.NewAccessToken(accessID, refreshID, input.UserID, domain.LoginID(input.LoginID), input.OrganizationName, now, accessExpiresAt)
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
	if err := c.addToWhitelist(ctx, c.refreshWhitelistRepo, input.UserID, domaintoken.WhitelistEntry{ID: refreshID, CreatedAt: now}); err != nil {
		return nil, fmt.Errorf("update refresh token whitelist: %w", err)
	}

	// TX4: Update access token whitelist
	if err := c.addToWhitelist(ctx, c.accessWhitelistRepo, input.UserID, domaintoken.WhitelistEntry{ID: accessID, CreatedAt: now}); err != nil {
		return nil, fmt.Errorf("update access token whitelist: %w", err)
	}

	output, err := authservice.NewCreateTokenPairOutput(jwtString, rawRefresh)
	if err != nil {
		return nil, fmt.Errorf("create token pair output: %w", err)
	}
	return output, nil
}

func (c *CreateTokenPairCommand) addToWhitelist(ctx context.Context, repo WhitelistRepository, userID int, entry domaintoken.WhitelistEntry) error {
	entries, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find whitelist entries: %w", err)
	}
	whitelist, err := domaintoken.NewWhitelist(userID, entries, c.config.TokenWhitelistSize)
	if err != nil {
		return fmt.Errorf("new whitelist: %w", err)
	}
	whitelist.Add(entry)
	if err := repo.Save(ctx, whitelist); err != nil {
		return fmt.Errorf("save whitelist: %w", err)
	}
	return nil
}
