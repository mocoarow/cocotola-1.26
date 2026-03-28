package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type validateSessionTokenRepo interface {
	FindByTokenHash(ctx context.Context, hash string) (*domaintoken.SessionToken, error)
}

// ValidateSessionTokenQuery validates a raw session token and returns user info.
type ValidateSessionTokenQuery struct {
	repo          validateSessionTokenRepo
	whitelistRepo whitelistFinder
	cache         sessionTokenCacheReadWriter
	config        UsecaseConfig
}

// NewValidateSessionTokenQuery returns a new ValidateSessionTokenQuery.
func NewValidateSessionTokenQuery(
	repo validateSessionTokenRepo,
	whitelistRepo whitelistFinder,
	cache sessionTokenCacheReadWriter,
	config UsecaseConfig,
) *ValidateSessionTokenQuery {
	return &ValidateSessionTokenQuery{
		repo:          repo,
		whitelistRepo: whitelistRepo,
		cache:         cache,
		config:        config,
	}
}

// ValidateSessionToken validates a raw session token and returns user info.
func (q *ValidateSessionTokenQuery) ValidateSessionToken(ctx context.Context, input *authservice.ValidateSessionTokenInput) (*authservice.ValidateSessionTokenOutput, error) {
	hash := string(domaintoken.HashToken(input.RawToken))
	now := q.config.Now()

	token, ok := q.cache.GetSessionToken(hash)
	if !ok {
		var err error
		token, err = q.repo.FindByTokenHash(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("find session token: %w", err)
		}

		// Check whitelist before caching
		entries, err := q.whitelistRepo.FindByUserID(ctx, token.UserID())
		if err != nil {
			return nil, fmt.Errorf("find session token whitelist: %w", err)
		}
		whitelist, err := domaintoken.NewWhitelist(token.UserID(), entries, q.config.TokenWhitelistSize)
		if err != nil {
			return nil, fmt.Errorf("new session token whitelist: %w", err)
		}
		if !whitelist.ContainsToken(token.ID()) {
			return nil, domain.ErrTokenNotFound
		}

		q.cache.SetSessionToken(hash, token)
	}

	if token.IsRevoked() {
		return nil, domain.ErrTokenRevoked
	}
	if token.IsExpired(now) {
		return nil, domain.ErrSessionExpired
	}
	maxTTL := time.Duration(q.config.SessionMaxTTLMin) * time.Minute
	if token.IsAbsoluteExpired(now, maxTTL) {
		return nil, domain.ErrSessionExpired
	}

	output, err := authservice.NewValidateSessionTokenOutput(token.UserID(), string(token.LoginID()), token.OrganizationName())
	if err != nil {
		return nil, fmt.Errorf("create validate session token output: %w", err)
	}
	return output, nil
}
