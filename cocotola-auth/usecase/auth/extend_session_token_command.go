package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type extendSessionTokenRepo interface {
	FindByTokenHash(ctx context.Context, hash string) (*domain.SessionToken, error)
	Save(ctx context.Context, token *domain.SessionToken) error
}

type extendSessionTokenWhitelistRepo interface {
	FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error)
}

type extendSessionTokenCache interface {
	GetSessionToken(hash string) (*domain.SessionToken, bool)
	SetSessionToken(hash string, token *domain.SessionToken)
}

// ExtendSessionTokenCommand extends the expiry of a session token (sliding window).
type ExtendSessionTokenCommand struct {
	repo          extendSessionTokenRepo
	whitelistRepo extendSessionTokenWhitelistRepo
	cache         extendSessionTokenCache
	config        AuthUsecaseConfig
}

// NewExtendSessionTokenCommand returns a new ExtendSessionTokenCommand.
func NewExtendSessionTokenCommand(
	repo extendSessionTokenRepo,
	whitelistRepo extendSessionTokenWhitelistRepo,
	cache extendSessionTokenCache,
	config AuthUsecaseConfig,
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
	hash := string(domain.HashToken(input.RawToken))
	now := c.config.Now()

	token, ok := c.cache.GetSessionToken(hash)
	if !ok {
		var err error
		token, err = c.repo.FindByTokenHash(ctx, hash)
		if err != nil {
			return fmt.Errorf("find session token: %w", err)
		}

		// Check whitelist before proceeding
		entries, err := c.whitelistRepo.FindByUserID(ctx, token.UserID())
		if err != nil {
			return fmt.Errorf("find session token whitelist: %w", err)
		}
		whitelist := domain.NewTokenWhitelist(token.UserID(), entries, c.config.TokenWhitelistSize)
		if !whitelist.ContainsToken(token.ID()) {
			return domain.ErrTokenNotFound
		}
	}

	if token.IsRevoked() {
		return domain.ErrTokenRevoked
	}
	if token.IsExpired(now) {
		return domain.ErrSessionExpired
	}
	maxTTL := time.Duration(c.config.SessionMaxTTLMin) * time.Minute
	if token.IsAbsoluteExpired(now, maxTTL) {
		return domain.ErrSessionExpired
	}

	slidingTTL := time.Duration(c.config.SessionTokenTTLMin) * time.Minute
	token.ExtendExpiry(now, slidingTTL, maxTTL)

	if err := c.repo.Save(ctx, token); err != nil {
		return fmt.Errorf("save session token: %w", err)
	}

	c.cache.SetSessionToken(hash, token)

	return nil
}
