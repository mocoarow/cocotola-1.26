// Package event handles domain event processing for the auth service.
package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type activeUserListRepository interface {
	FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveUserList, error)
	Save(ctx context.Context, list *domain.ActiveUserList) error
}

type organizationFinder interface {
	FindByID(ctx context.Context, id int) (*domain.Organization, error)
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

	org, err := h.orgRepo.FindByID(ctx, e.OrganizationID)
	if err != nil {
		return fmt.Errorf("find organization %d: %w", e.OrganizationID, err)
	}

	activeUserList, err := h.activeUserRepo.FindByOrganizationID(ctx, e.OrganizationID)
	if err != nil {
		return fmt.Errorf("find active user list: %w", err)
	}

	if err := activeUserList.Add(e.AppUserID, org.MaxActiveUsers()); err != nil {
		return fmt.Errorf("add to active user list: %w", err)
	}

	if err := h.activeUserRepo.Save(ctx, activeUserList); err != nil {
		return fmt.Errorf("save active user list: %w", err)
	}

	h.logger.InfoContext(ctx, "added user to active user list",
		slog.Int("app_user_id", e.AppUserID),
		slog.Int("organization_id", e.OrganizationID))

	return nil
}
