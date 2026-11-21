package auth

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"
)

func updateWhitelist(ctx context.Context, repo WhitelistRepository, userID domain.AppUserID, mutate func(*domaintoken.Whitelist) error) error {
	if err := occ.Retry(func() error {
		whitelist, err := repo.FindByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("find: %w", err)
		}
		if err := mutate(whitelist); err != nil {
			return fmt.Errorf("mutate: %w", err)
		}
		if err := repo.Save(ctx, whitelist); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("update whitelist of user %s: %w", userID, err)
	}
	return nil
}
