package auth

import (
	"context"
	"fmt"
	"time"

	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// ExtendSessionTokenCommand extends the expiry of a session token (sliding window).
type ExtendSessionTokenCommand struct {
	repo          SessionTokenRepository
	whitelistRepo whitelistFinder
	cache         sessionTokenCacheReadWriter
	config        UsecaseConfig
}

// NewExtendSessionTokenCommand returns a new ExtendSessionTokenCommand.
func NewExtendSessionTokenCommand(
	repo SessionTokenRepository,
	whitelistRepo whitelistFinder,
	cache sessionTokenCacheReadWriter,
	config UsecaseConfig,
) *ExtendSessionTokenCommand {
	return &ExtendSessionTokenCommand{
		repo:          repo,
		whitelistRepo: whitelistRepo,
		cache:         cache,
		config:        config,
	}
}

// ExtendSessionToken extends the expiry of a session token (sliding window).
func (c *ExtendSessionTokenCommand) ExtendSessionToken(ctx context.Context, input *authservice.ExtendSessionTokenInput) error {
	hash := string(domaintoken.HashToken(input.RawToken))
	now := c.config.Now()

	token, ok := c.cache.GetSessionToken(hash)
	if !ok {
		var err error
		token, err = findWhitelistedSessionToken(ctx, c.repo, c.whitelistRepo, c.config.TokenWhitelistSize, hash)
		if err != nil {
			return err
		}
	}

	maxTTL := time.Duration(c.config.SessionMaxTTLMin) * time.Minute
	if err := checkSessionTokenUsable(token, now, maxTTL); err != nil {
		return err
	}

	slidingTTL := time.Duration(c.config.SessionTokenTTLMin) * time.Minute
	token.ExtendExpiry(now, slidingTTL, maxTTL)

	if err := c.repo.Save(ctx, token); err != nil {
		return fmt.Errorf("save session token: %w", err)
	}

	c.cache.SetSessionToken(hash, token)

	return nil
}
