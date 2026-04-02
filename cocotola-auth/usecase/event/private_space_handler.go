package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
)

type spaceCreator interface {
	Create(ctx context.Context, organizationID int, ownerID int, keyName string, name string, spaceType string, createdBy int) (int, error)
}

type userPolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID int, userID int, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error
}

// PrivateSpaceHandler creates a private space when a new app user is created.
type PrivateSpaceHandler struct {
	spaceRepo  spaceCreator
	policyRepo userPolicyAdder
	logger     *slog.Logger
}

// NewPrivateSpaceHandler returns a new PrivateSpaceHandler.
func NewPrivateSpaceHandler(
	spaceRepo spaceCreator,
	policyRepo userPolicyAdder,
	logger *slog.Logger,
) *PrivateSpaceHandler {
	return &PrivateSpaceHandler{
		spaceRepo:  spaceRepo,
		policyRepo: policyRepo,
		logger:     logger,
	}
}

// Handle processes an AppUserCreated event by creating a private space for the user.
func (h *PrivateSpaceHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(domain.AppUserCreated)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	keyName := domainspace.PrivateSpaceKeyName(e.LoginID())
	spaceName := fmt.Sprintf("Private(%s)", e.LoginID())

	spaceID, err := h.spaceRepo.Create(ctx, e.OrganizationID(), e.AppUserID(), keyName, spaceName, domainspace.TypePrivate().Value(), e.AppUserID())
	if err != nil {
		return fmt.Errorf("create private space for user %d: %w", e.AppUserID(), err)
	}

	if err := h.policyRepo.AddPolicyForUser(ctx, e.OrganizationID(), e.AppUserID(), domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
		return fmt.Errorf("add list_spaces policy for user %d: %w", e.AppUserID(), err)
	}

	if err := h.policyRepo.AddPolicyForUser(ctx, e.OrganizationID(), e.AppUserID(), domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(spaceID), domainrbac.EffectAllow()); err != nil {
		return fmt.Errorf("add view_space policy for user %d on space %d: %w", e.AppUserID(), spaceID, err)
	}

	workbookActions := []domainrbac.Action{
		domainrbac.ActionCreateWorkbook(),
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionUpdateWorkbook(),
		domainrbac.ActionDeleteWorkbook(),
		domainrbac.ActionImportWorkbook(),
	}
	for _, action := range workbookActions {
		if err := h.policyRepo.AddPolicyForUser(ctx, e.OrganizationID(), e.AppUserID(), action, domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add %s policy for user %d: %w", action.Value(), e.AppUserID(), err)
		}
	}

	h.logger.InfoContext(ctx, "private space created for user",
		slog.Int("user_id", e.AppUserID()),
		slog.Int("space_id", spaceID),
		slog.Int("organization_id", e.OrganizationID()),
	)

	return nil
}
