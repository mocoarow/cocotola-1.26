package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

func findWhitelistedSessionToken(
	ctx context.Context,
	repo sessionTokenFinder,
	whitelistRepo whitelistFinder,
	hash string,
) (*domaintoken.SessionToken, error) {
	token, err := repo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("find session token: %w", err)
	}

	whitelist, err := whitelistRepo.FindByUserID(ctx, token.UserID())
	if err != nil {
		return nil, fmt.Errorf("find session token whitelist: %w", err)
	}
	if !whitelist.ContainsToken(token.ID()) {
		return nil, domain.ErrTokenNotFound
	}

	return token, nil
}

func checkSessionTokenUsable(token *domaintoken.SessionToken, now time.Time, maxTTL time.Duration) error {
	if token.IsRevoked() {
		return domain.ErrTokenRevoked
	}
	if token.IsExpired(now) || token.IsAbsoluteExpired(now, maxTTL) {
		return domain.ErrSessionExpired
	}
	return nil
}
