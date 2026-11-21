package auth

import (
	"context"
	"fmt"
	"time"

	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// ValidateSessionTokenQuery validates a raw session token and returns user info.
type ValidateSessionTokenQuery struct {
	repo          sessionTokenFinder
	whitelistRepo whitelistFinder
	cache         sessionTokenCacheReadWriter
	config        UsecaseConfig
}

// NewValidateSessionTokenQuery returns a new ValidateSessionTokenQuery.
func NewValidateSessionTokenQuery(
	repo sessionTokenFinder,
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
		token, err = findWhitelistedSessionToken(ctx, q.repo, q.whitelistRepo, hash)
		if err != nil {
			return nil, err
		}
		q.cache.SetSessionToken(hash, token)
	}

	maxTTL := time.Duration(q.config.SessionMaxTTLMin) * time.Minute
	if err := checkSessionTokenUsable(token, now, maxTTL); err != nil {
		return nil, err
	}

	output, err := authservice.NewValidateSessionTokenOutput(token.UserID(), string(token.LoginID()), token.OrganizationName())
	if err != nil {
		return nil, fmt.Errorf("create validate session token output: %w", err)
	}
	return output, nil
}
