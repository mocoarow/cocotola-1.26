package auth

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type revokeSessionTokenCache interface {
	GetSessionToken(hash string) (*domaintoken.SessionToken, bool)
	DeleteSessionToken(hash string)
}

// RevokeSessionTokenCommand revokes a session token.
type RevokeSessionTokenCommand struct {
	repo          SessionTokenRepository
	whitelistRepo WhitelistRepository
	cache         revokeSessionTokenCache
	config        UsecaseConfig
}

// NewRevokeSessionTokenCommand returns a new RevokeSessionTokenCommand.
func NewRevokeSessionTokenCommand(
	repo SessionTokenRepository,
	whitelistRepo WhitelistRepository,
	cache revokeSessionTokenCache,
	config UsecaseConfig,
) *RevokeSessionTokenCommand {
	return &RevokeSessionTokenCommand{
		repo:          repo,
		whitelistRepo: whitelistRepo,
		cache:         cache,
		config:        config,
	}
}

// RevokeSessionToken revokes a session token.
func (c *RevokeSessionTokenCommand) RevokeSessionToken(ctx context.Context, input *authservice.RevokeSessionTokenInput) error {
	hash := string(domaintoken.HashToken(input.RawToken))

	token, ok := c.cache.GetSessionToken(hash)
	if !ok {
		var err error
		token, err = c.repo.FindByTokenHash(ctx, hash)
		if err != nil {
			return fmt.Errorf("find session token: %w", err)
		}
	}

	if token.IsRevoked() {
		return domain.ErrTokenRevoked
	}

	// TX1: Revoke session token
	now := c.config.Now()
	token.Revoke(now)
	if err := c.repo.Save(ctx, token); err != nil {
		return fmt.Errorf("save session token: %w", err)
	}

	c.cache.DeleteSessionToken(hash)

	// TX2: Remove from whitelist (separate aggregate, separate transaction)
	entries, err := c.whitelistRepo.FindByUserID(ctx, token.UserID())
	if err != nil {
		return fmt.Errorf("find session token whitelist: %w", err)
	}

	whitelist, err := domaintoken.NewWhitelist(token.UserID(), entries, c.config.TokenWhitelistSize)
	if err != nil {
		return fmt.Errorf("new session token whitelist: %w", err)
	}
	whitelist.Remove([]string{token.ID()})

	if err := c.whitelistRepo.Save(ctx, whitelist); err != nil {
		return fmt.Errorf("save session token whitelist: %w", err)
	}

	return nil
}
