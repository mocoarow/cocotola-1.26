package event

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
)

type spaceSaver interface {
	Save(ctx context.Context, space *domainspace.Space) error
}

type publicSpaceFinder interface {
	FindPublicByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*domainspace.Space, error)
}

type userPolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error
}

// PrivateSpaceHandler creates a private space when a new app user is created.
type PrivateSpaceHandler struct {
	spaceRepo         spaceSaver
	publicSpaceFinder publicSpaceFinder
	policyRepo        userPolicyAdder
	logger            *slog.Logger
}

// NewPrivateSpaceHandler returns a new PrivateSpaceHandler.
func NewPrivateSpaceHandler(
	spaceRepo spaceSaver,
	publicSpaceFinder publicSpaceFinder,
	policyRepo userPolicyAdder,
	logger *slog.Logger,
) *PrivateSpaceHandler {
	return &PrivateSpaceHandler{
		spaceRepo:         spaceRepo,
		publicSpaceFinder: publicSpaceFinder,
		policyRepo:        policyRepo,
		logger:            logger,
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

	spaceResource := domainrbac.ResourceSpace(s.ID())
	spaceWorkbookActions := []domainrbac.Action{
		domainrbac.ActionCreateWorkbook(),
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionUpdateWorkbook(),
		domainrbac.ActionDeleteWorkbook(),
	}
	for _, action := range spaceWorkbookActions {
		if err := h.policyRepo.AddPolicyForUser(ctx, orgID, userID, action, spaceResource, domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add %s policy for user %s: %w", action.Value(), userID.String(), err)
		}
	}

	if err := h.policyRepo.AddPolicyForUser(ctx, orgID, userID, domainrbac.ActionImportWorkbook(), domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
		return fmt.Errorf("add %s policy for user %s: %w", domainrbac.ActionImportWorkbook().Value(), userID.String(), err)
	}

	if err := h.grantPublicSpacePolicies(ctx, orgID, userID); err != nil {
		return err
	}

	h.logger.InfoContext(ctx, "private space created for user",
		slog.String("user_id", userID.String()),
		slog.String("space_id", s.ID().String()),
		slog.String("organization_id", orgID.String()),
	)

	return nil
}

// grantPublicSpacePolicies grants view_workbook and create_workbook on the public space.
func (h *PrivateSpaceHandler) grantPublicSpacePolicies(ctx context.Context, orgID domain.OrganizationID, userID domain.AppUserID) error {
	publicSpace, err := h.publicSpaceFinder.FindPublicByOrganizationID(ctx, orgID)
	if err != nil {
		if errors.Is(err, domain.ErrSpaceNotFound) {
			h.logger.InfoContext(ctx, "public space not found, skipping public space policies",
				slog.String("organization_id", orgID.String()),
			)
			return nil
		}
		return fmt.Errorf("find public space for organization %s: %w", orgID.String(), err)
	}

	publicResource := domainrbac.ResourceSpace(publicSpace.ID())
	publicActions := []domainrbac.Action{
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionCreateWorkbook(),
	}
	for _, action := range publicActions {
		if err := h.policyRepo.AddPolicyForUser(ctx, orgID, userID, action, publicResource, domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add %s policy on public space for user %s: %w", action.Value(), userID.String(), err)
		}
	}
	return nil
}
