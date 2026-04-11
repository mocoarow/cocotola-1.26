package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
)

type spaceSaver interface {
	Save(ctx context.Context, space *domainspace.Space) error
}

type userPolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error
}

// PrivateSpaceHandler creates a private space when a new app user is created.
type PrivateSpaceHandler struct {
	spaceRepo  spaceSaver
	policyRepo userPolicyAdder
	logger     *slog.Logger
}

// NewPrivateSpaceHandler returns a new PrivateSpaceHandler.
func NewPrivateSpaceHandler(
	spaceRepo spaceSaver,
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
		return fmt.Errorf("unexpected event type %T: %w", event, domain.ErrInvalidArgument)
	}

	keyName := domainspace.PrivateSpaceKeyName(e.LoginID())
	spaceName := fmt.Sprintf("Private(%s)", e.LoginID())

	orgID := e.OrganizationID()
	userID := e.AppUserID()

	s, err := domainspace.Provision(ctx, h.spaceRepo, orgID, userID, keyName, spaceName, domainspace.TypePrivate())
	if err != nil {
		return fmt.Errorf("provision private space for user %s: %w", userID.String(), err)
	}

	if err := h.policyRepo.AddPolicyForUser(ctx, orgID, userID, domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
		return fmt.Errorf("add list_spaces policy for user %s: %w", userID.String(), err)
	}

	if err := h.policyRepo.AddPolicyForUser(ctx, orgID, userID, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(s.ID()), domainrbac.EffectAllow()); err != nil {
		return fmt.Errorf("add view_space policy for user %s on space %s: %w", userID.String(), s.ID().String(), err)
	}

	workbookActions := []domainrbac.Action{
		domainrbac.ActionCreateWorkbook(),
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionUpdateWorkbook(),
		domainrbac.ActionDeleteWorkbook(),
		domainrbac.ActionImportWorkbook(),
	}
	for _, action := range workbookActions {
		if err := h.policyRepo.AddPolicyForUser(ctx, orgID, userID, action, domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add %s policy for user %s: %w", action.Value(), userID.String(), err)
		}
	}

	h.logger.InfoContext(ctx, "private space created for user",
		slog.String("user_id", userID.String()),
		slog.String("space_id", s.ID().String()),
		slog.String("organization_id", orgID.String()),
	)

	return nil
}
