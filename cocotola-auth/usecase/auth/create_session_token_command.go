package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type createSessionTokenRepo interface {
	Save(ctx context.Context, token *domain.SessionToken) error
}

type createSessionTokenCache interface {
	SetSessionToken(hash string, token *domain.SessionToken)
}

// CreateSessionTokenCommand creates a new session token for cookie-based authentication.
type CreateSessionTokenCommand struct {
	repo          createSessionTokenRepo
	whitelistRepo WhitelistRepository
	cache         createSessionTokenCache
	config        UsecaseConfig
}

// NewCreateSessionTokenCommand returns a new CreateSessionTokenCommand.
func NewCreateSessionTokenCommand(
	repo createSessionTokenRepo,
	whitelistRepo WhitelistRepository,
	cache createSessionTokenCache,
	config UsecaseConfig,
) *CreateSessionTokenCommand {
	return &CreateSessionTokenCommand{
		repo:          repo,
		whitelistRepo: whitelistRepo,
		cache:         cache,
		config:        config,
	}
}

// CreateSessionToken generates an opaque session token, persists it, and returns the raw token.
func (c *CreateSessionTokenCommand) CreateSessionToken(ctx context.Context, input *authservice.CreateSessionTokenInput) (*authservice.CreateSessionTokenOutput, error) {
	raw, hash, err := domain.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate opaque token: %w", err)
	}

	now := c.config.Now()
	expiresAt := now.Add(time.Duration(c.config.SessionTokenTTLMin) * time.Minute)
	id := uuid.New().String()

	token, err := domain.NewSessionToken(id, input.UserID, domain.LoginID(input.LoginID), input.OrganizationName, hash, now, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("new session token: %w", err)
	}

	// TX1: Save session token
	if err := c.repo.Save(ctx, token); err != nil {
		return nil, fmt.Errorf("save session token: %w", err)
	}

	c.cache.SetSessionToken(string(hash), token)

	// TX2: Update whitelist (separate aggregate, separate transaction)
	entries, err := c.whitelistRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("find session token whitelist entries: %w", err)
	}

	whitelist := domain.NewTokenWhitelist(input.UserID, entries, c.config.TokenWhitelistSize)
	whitelist.Add(domain.WhitelistEntry{ID: id, CreatedAt: now})

	if err := c.whitelistRepo.Save(ctx, whitelist); err != nil {
		return nil, fmt.Errorf("save session token whitelist: %w", err)
	}

	output, err := authservice.NewCreateSessionTokenOutput(raw)
	if err != nil {
		return nil, fmt.Errorf("create session token output: %w", err)
	}
	return output, nil
}
