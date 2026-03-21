package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type activeGroupListRepository interface {
	FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveGroupList, error)
	Save(ctx context.Context, list *domain.ActiveGroupList) error
}

// ActiveGroupListHandler adds a newly created group to the organization's active group list.
type ActiveGroupListHandler struct {
	activeGroupRepo activeGroupListRepository
	orgRepo         organizationFinder
	logger          *slog.Logger
}

// NewActiveGroupListHandler returns a new ActiveGroupListHandler.
func NewActiveGroupListHandler(
	activeGroupRepo activeGroupListRepository,
	orgRepo organizationFinder,
	logger *slog.Logger,
) *ActiveGroupListHandler {
	return &ActiveGroupListHandler{
		activeGroupRepo: activeGroupRepo,
		orgRepo:         orgRepo,
		logger:          logger,
	}
}

// Handle processes a GroupCreated event by adding the group to the active group list.
func (h *ActiveGroupListHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(domain.GroupCreated)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	return handleActiveListEvent[domain.ActiveGroupList](
		ctx, h.orgRepo, h.activeGroupRepo,
		e.OrganizationID, e.GroupID,
		func(org *domain.Organization) int { return org.MaxActiveGroups() },
		"group", h.logger,
	)
}
