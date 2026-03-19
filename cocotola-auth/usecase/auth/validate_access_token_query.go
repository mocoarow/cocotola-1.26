package auth

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type validateAccessTokenRepo interface {
	FindByID(ctx context.Context, id string) (*domain.AccessToken, error)
}

type validateAccessTokenCache interface {
	GetAccessToken(jti string) (*domain.AccessToken, bool)
	SetAccessToken(jti string, token *domain.AccessToken)
	DeleteAccessToken(jti string)
}

// ValidateAccessTokenQuery validates a JWT access token against the whitelist.
type ValidateAccessTokenQuery struct {
	repo          validateAccessTokenRepo
	whitelistRepo whitelistFinder
	jwt           jwtParser
	cache         validateAccessTokenCache
	config        UsecaseConfig
}

// NewValidateAccessTokenQuery returns a new ValidateAccessTokenQuery.
func NewValidateAccessTokenQuery(
	repo validateAccessTokenRepo,
	whitelistRepo whitelistFinder,
	jwt jwtParser,
	cache validateAccessTokenCache,
	config UsecaseConfig,
) *ValidateAccessTokenQuery {
	return &ValidateAccessTokenQuery{
		repo:          repo,
		whitelistRepo: whitelistRepo,
		jwt:           jwt,
		cache:         cache,
		config:        config,
	}
}

// ValidateAccessToken validates a JWT access token against the whitelist.
func (q *ValidateAccessTokenQuery) ValidateAccessToken(ctx context.Context, input *authservice.ValidateAccessTokenInput) (*authservice.ValidateAccessTokenOutput, error) {
	userInfo, jti, err := q.jwt.ParseAccessToken(input.JWTString)
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	now := q.config.Now()

	token, ok := q.cache.GetAccessToken(jti)
	if !ok {
		// Check whitelist before loading full token
		entries, err := q.whitelistRepo.FindByUserID(ctx, userInfo.UserID)
		if err != nil {
			return nil, fmt.Errorf("find access token whitelist: %w", err)
		}
		whitelist := domain.NewTokenWhitelist(userInfo.UserID, entries, q.config.TokenWhitelistSize)
		if !whitelist.ContainsToken(jti) {
			return nil, domain.ErrTokenNotFound
		}

		token, err = q.repo.FindByID(ctx, jti)
		if err != nil {
			return nil, fmt.Errorf("find access token: %w", err)
		}
	}

	if token.IsRevoked() {
		q.cache.DeleteAccessToken(jti)
		return nil, domain.ErrTokenRevoked
	}

	if !ok {
		q.cache.SetAccessToken(jti, token)
	}
	if token.IsExpired(now) {
		return nil, domain.ErrSessionExpired
	}

	output, err := authservice.NewValidateAccessTokenOutput(userInfo.UserID, userInfo.LoginID, userInfo.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("create validate access token output: %w", err)
	}
	return output, nil
}
