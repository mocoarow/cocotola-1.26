package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"
)

// activeList is the interface that both ActiveUserList and ActiveGroupList satisfy.
type activeList[T any, E any] interface {
	*T
	Add(id E, limit int) error
}

// activeListRepo is the interface for retrieving and saving an active list.
type activeListRepo[T any] interface {
	FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*T, error)
	Save(ctx context.Context, list *T) error
}

// handleActiveListEvent processes an event that adds an entry to an active list.
func handleActiveListEvent[T any, E any, L activeList[T, E]](
	ctx context.Context,
	orgRepo organizationFinder,
	repo activeListRepo[T],
	organizationID domain.OrganizationID,
	entityID E,
	maxFn func(*domain.Organization) int,
	entityLabel string,
	logger *slog.Logger,
) error {
	org, err := orgRepo.FindByID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("find organization %s: %w", organizationID.String(), err)
	}

	if err := occ.Retry(func() error {
		list, err := repo.FindByOrganizationID(ctx, organizationID)
		if err != nil {
			return fmt.Errorf("find: %w", err)
		}
		if err := L(list).Add(entityID, maxFn(org)); err != nil {
			return fmt.Errorf("add: %w", err)
		}
		if err := repo.Save(ctx, list); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("update active %s list of organization %s: %w", entityLabel, organizationID, err)
	}

	logger.InfoContext(ctx, fmt.Sprintf("added %s to active %s list", entityLabel, entityLabel),
		slog.Any(entityLabel+"_id", entityID),
		slog.String("organization_id", organizationID.String()))

	return nil
}
