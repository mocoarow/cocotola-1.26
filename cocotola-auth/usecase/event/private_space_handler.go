package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type spaceCreator interface {
	Create(ctx context.Context, organizationID int, ownerID int, keyName string, name string, spaceType string, createdBy int) (int, error)
}

type userSpaceAdder interface {
	AddUserToSpace(ctx context.Context, organizationID int, userID int, spaceID int, createdBy int) error
}

// PrivateSpaceHandler creates a private space when a new app user is created.
type PrivateSpaceHandler struct {
	spaceRepo     spaceCreator
	userSpaceRepo userSpaceAdder
	logger        *slog.Logger
}

// NewPrivateSpaceHandler returns a new PrivateSpaceHandler.
func NewPrivateSpaceHandler(
	spaceRepo spaceCreator,
	userSpaceRepo userSpaceAdder,
	logger *slog.Logger,
) *PrivateSpaceHandler {
	return &PrivateSpaceHandler{
		spaceRepo:     spaceRepo,
		userSpaceRepo: userSpaceRepo,
		logger:        logger,
	}
}

// Handle processes an AppUserCreated event by creating a private space for the user.
func (h *PrivateSpaceHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(domain.AppUserCreated)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	keyName := domain.PrivateSpaceKeyName(e.LoginID)
	spaceName := fmt.Sprintf("Private(%s)", e.LoginID)

	spaceID, err := h.spaceRepo.Create(ctx, e.OrganizationID, e.AppUserID, keyName, spaceName, domain.SpaceTypePrivate().Value(), e.AppUserID)
	if err != nil {
		return fmt.Errorf("create private space for user %d: %w", e.AppUserID, err)
	}

	if err := h.userSpaceRepo.AddUserToSpace(ctx, e.OrganizationID, e.AppUserID, spaceID, e.AppUserID); err != nil {
		return fmt.Errorf("add user %d to private space %d: %w", e.AppUserID, spaceID, err)
	}

	h.logger.InfoContext(ctx, "private space created for user",
		slog.Int("user_id", e.AppUserID),
		slog.Int("space_id", spaceID),
		slog.Int("organization_id", e.OrganizationID),
	)

	return nil
}
