package workbook

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// CreateWorkbookCommand handles workbook creation.
type CreateWorkbookCommand struct {
	workbookRepo     workbookCreator
	ownedListFinder  ownedWorkbookListFinder
	ownedListSaver   ownedWorkbookListSaver
	maxWbFetcher     maxWorkbooksFetcher
	spaceTypeFetcher spaceTypeFetcher
	authChecker      authorizationChecker
	policyAdder      policyAdder
}

// NewCreateWorkbookCommand returns a new CreateWorkbookCommand.
func NewCreateWorkbookCommand(
	workbookRepo workbookCreator,
	ownedListFinder ownedWorkbookListFinder,
	ownedListSaver ownedWorkbookListSaver,
	maxWbFetcher maxWorkbooksFetcher,
	spaceTypeFetcher spaceTypeFetcher,
	authChecker authorizationChecker,
	policyAdder policyAdder,
) *CreateWorkbookCommand {
	return &CreateWorkbookCommand{
		workbookRepo:     workbookRepo,
		ownedListFinder:  ownedListFinder,
		ownedListSaver:   ownedListSaver,
		maxWbFetcher:     maxWbFetcher,
		spaceTypeFetcher: spaceTypeFetcher,
		authChecker:      authChecker,
		policyAdder:      policyAdder,
	}
}

// CreateWorkbook creates a new workbook.
//
// The caller-supplied Visibility is overwritten to match the SpaceType of the
// target space: PublicSpace ⇒ "public", PrivateSpace ⇒ "private". This keeps
// the dataset internally consistent regardless of what an authenticated client
// (or a buggy admin tool) sends in.
func (c *CreateWorkbookCommand) CreateWorkbook(ctx context.Context, input *workbookservice.CreateWorkbookInput) (*workbookservice.CreateWorkbookOutput, error) {
	if err := c.authorizeCreateWorkbook(ctx, input); err != nil {
		return nil, fmt.Errorf("authorize create workbook: %w", err)
	}

	visibility, err := c.resolveVisibility(ctx, input.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("resolve visibility: %w", err)
	}

	// Validate language via the domain value object so anything that slips
	// past the service-layer length check (e.g. "JA", "あa") is rejected
	// before it reaches Firestore.
	lang, err := domainworkbook.NewLanguage(input.Language)
	if err != nil {
		return nil, fmt.Errorf("new language: %w", err)
	}

	ownedList, maxWorkbooks, err := c.loadOwnedListWithLimit(ctx, input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("load owned list with limit: %w", err)
	}

	workbookID, err := c.workbookRepo.Create(ctx, input.SpaceID, input.OperatorID, input.OrganizationID, input.Title, input.Description, visibility, lang.Value())
	if err != nil {
		return nil, fmt.Errorf("create workbook: %w", err)
	}

	if err := c.grantWorkbookPolicies(ctx, input.OrganizationID, input.OperatorID, workbookID); err != nil {
		return nil, fmt.Errorf("grant workbook policies: %w", err)
	}

	if err := c.saveOwnedList(ctx, ownedList, workbookID, input.OperatorID, maxWorkbooks); err != nil {
		return nil, fmt.Errorf("save owned list: %w", err)
	}

	now := time.Now()
	output, err := workbookservice.NewCreateWorkbookOutput(workbookID, input.SpaceID, input.OperatorID, input.OrganizationID, input.Title, input.Description, visibility, lang.Value(), now, now)
	if err != nil {
		return nil, fmt.Errorf("create workbook output: %w", err)
	}
	return output, nil
}

// resolveVisibility consults cocotola-auth for the SpaceType of the target space
// and maps it to the canonical Workbook visibility ("public" / "private"). This
// is the single source of truth for visibility — clients cannot override it.
func (c *CreateWorkbookCommand) resolveVisibility(ctx context.Context, spaceID string) (string, error) {
	spaceType, err := c.spaceTypeFetcher.FetchSpaceType(ctx, spaceID)
	if err != nil {
		return "", fmt.Errorf("fetch space type for space %s: %w", spaceID, err)
	}

	switch spaceType {
	case "public":
		return domainworkbook.VisibilityPublic().Value(), nil
	case "private":
		return domainworkbook.VisibilityPrivate().Value(), nil
	default:
		return "", fmt.Errorf("unsupported space type %q: %w", spaceType, domain.ErrInvalidArgument)
	}
}

func (c *CreateWorkbookCommand) authorizeCreateWorkbook(ctx context.Context, input *workbookservice.CreateWorkbookInput) error {
	spaceResource, err := domain.ResourceSpace(input.SpaceID)
	if err != nil {
		return fmt.Errorf("resource space: %w", err)
	}
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionCreateWorkbook(), spaceResource)
	if err != nil {
		return fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

func (c *CreateWorkbookCommand) loadOwnedListWithLimit(ctx context.Context, operatorID string) (*domain.OwnedWorkbookList, int, error) {
	ownedList, err := c.ownedListFinder.FindByOwnerID(ctx, operatorID)
	if err != nil {
		return nil, 0, fmt.Errorf("find owned workbook list: %w", err)
	}

	maxWorkbooks, err := c.maxWbFetcher.FetchMaxWorkbooks(ctx, operatorID)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch max workbooks: %w", err)
	}

	if ownedList.Size() >= maxWorkbooks {
		return nil, 0, domain.ErrOwnedWorkbookLimitReached
	}

	return ownedList, maxWorkbooks, nil
}

// saveOwnedList adds the workbook to the owned list and persists it.
// NOTE: eventual consistency -- if Save fails, the workbook exists but is not tracked
// in the owned list. This is by design: OwnedWorkbookList and Workbook are separate
// aggregates, so cross-aggregate consistency is eventual. A periodic reconciliation
// process can detect and resolve orphaned workbooks.
func (c *CreateWorkbookCommand) saveOwnedList(ctx context.Context, ownedList *domain.OwnedWorkbookList, workbookID, operatorID string, maxWorkbooks int) error {
	if err := ownedList.Add(workbookID, maxWorkbooks); err != nil {
		return fmt.Errorf("add to owned workbook list: %w", err)
	}
	if err := c.ownedListSaver.Save(ctx, ownedList); err != nil {
		slog.ErrorContext(ctx, "owned list save failed after workbook creation",
			slog.String("workbook_id", workbookID),
			slog.String("owner_id", operatorID),
			slog.Any("error", err),
		)
		return fmt.Errorf("save owned workbook list: %w", err)
	}
	return nil
}

// grantWorkbookPolicies grants workbook-scoped permissions via RBAC.
// NOTE: eventual consistency -- if a policy grant fails partway through, some policies
// will exist and others will not. This is acceptable because cocotola-auth is a separate
// service (cross-service consistency is eventual). A retry or reconciliation mechanism
// can re-grant missing policies.
func (c *CreateWorkbookCommand) grantWorkbookPolicies(ctx context.Context, organizationID, operatorID, workbookID string) error {
	actions := []domain.Action{
		domain.ActionViewWorkbook(),
		domain.ActionUpdateWorkbook(),
		domain.ActionDeleteWorkbook(),
		domain.ActionCreateQuestion(),
		domain.ActionUpdateQuestion(),
		domain.ActionDeleteQuestion(),
	}
	resource, err := domain.ResourceWorkbook(workbookID)
	if err != nil {
		return fmt.Errorf("resource workbook: %w", err)
	}
	for _, action := range actions {
		if err := c.policyAdder.AddPolicyForUser(ctx, organizationID, operatorID, action, resource, domain.EffectAllow()); err != nil {
			return fmt.Errorf("add %s policy for workbook %s: %w", action.Value(), workbookID, err)
		}
	}
	return nil
}
