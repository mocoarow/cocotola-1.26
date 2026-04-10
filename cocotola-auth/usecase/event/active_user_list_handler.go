// Package event handles domain event processing for the auth service.
package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type activeUserListRepository interface {
	FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*domain.ActiveUserList, error)
	Save(ctx context.Context, list *domain.ActiveUserList) error
}

type organizationFinder interface {
	FindByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
}

// ActiveUserListHandler adds a newly created user to the organization's active user list.
type ActiveUserListHandler struct {
	activeUserRepo activeUserListRepository
	orgRepo        organizationFinder
	logger         *slog.Logger
}

// NewActiveUserListHandler returns a new ActiveUserListHandler.
func NewActiveUserListHandler(
	activeUserRepo activeUserListRepository,
	orgRepo organizationFinder,
	logger *slog.Logger,
) *ActiveUserListHandler {
	return &ActiveUserListHandler{
		activeUserRepo: activeUserRepo,
		orgRepo:        orgRepo,
		logger:         logger,
	}
}

// Handle processes an AppUserCreated event by adding the user to the active user list.
func (h *ActiveUserListHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(domain.AppUserCreated)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	return handleActiveListEvent[domain.ActiveUserList, domain.AppUserID](
		ctx, h.orgRepo, h.activeUserRepo,
		e.OrganizationID(), e.AppUserID(),
		func(org *domain.Organization) int { return org.MaxActiveUsers() },
		"user", h.logger,
	)
}
