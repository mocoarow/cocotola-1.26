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
	workbookRepo    workbookCreator
	ownedListFinder ownedWorkbookListFinder
	ownedListSaver  ownedWorkbookListSaver
	maxWbFetcher    maxWorkbooksFetcher
	authChecker     authorizationChecker
	policyAdder     policyAdder
}

// NewCreateWorkbookCommand returns a new CreateWorkbookCommand.
func NewCreateWorkbookCommand(
	workbookRepo workbookCreator,
	ownedListFinder ownedWorkbookListFinder,
	ownedListSaver ownedWorkbookListSaver,
	maxWbFetcher maxWorkbooksFetcher,
	authChecker authorizationChecker,
	policyAdder policyAdder,
) *CreateWorkbookCommand {
	return &CreateWorkbookCommand{
		workbookRepo:    workbookRepo,
		ownedListFinder: ownedListFinder,
		ownedListSaver:  ownedListSaver,
		maxWbFetcher:    maxWbFetcher,
		authChecker:     authChecker,
		policyAdder:     policyAdder,
	}
}

// CreateWorkbook creates a new workbook.
func (c *CreateWorkbookCommand) CreateWorkbook(ctx context.Context, input *workbookservice.CreateWorkbookInput) (*workbookservice.CreateWorkbookOutput, error) {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	if _, err := domainworkbook.NewVisibility(input.Visibility); err != nil {
		return nil, fmt.Errorf("new visibility: %w", err)
	}

	// Load owned workbook list and check limit before creating.
	ownedList, err := c.ownedListFinder.FindByOwnerID(ctx, input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("find owned workbook list: %w", err)
	}

	maxWorkbooks, err := c.maxWbFetcher.FetchMaxWorkbooks(ctx, input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("fetch max workbooks: %w", err)
	}

	// Pre-check capacity before creating the workbook.
	if ownedList.Size() >= maxWorkbooks {
		return nil, domain.ErrOwnedWorkbookLimitReached
	}

	workbookID, err := c.workbookRepo.Create(ctx, input.SpaceID, input.OperatorID, input.OrganizationID, input.Title, input.Description, input.Visibility)
	if err != nil {
		return nil, fmt.Errorf("create workbook: %w", err)
	}

	if err := c.grantWorkbookPolicies(ctx, input.OrganizationID, input.OperatorID, workbookID); err != nil {
		return nil, err
	}

	// Add to owned list and persist (optimistic lock via version).
	// NOTE: eventual consistency -- if Save fails, the workbook exists but is not tracked
	// in the owned list. This is by design: OwnedWorkbookList and Workbook are separate
	// aggregates, so cross-aggregate consistency is eventual. A periodic reconciliation
	// process can detect and resolve orphaned workbooks.
	if err := ownedList.Add(workbookID, maxWorkbooks); err != nil {
		return nil, fmt.Errorf("add to owned workbook list: %w", err)
	}
	if err := c.ownedListSaver.Save(ctx, ownedList); err != nil {
		slog.ErrorContext(ctx, "owned list save failed after workbook creation",
			slog.String("workbook_id", workbookID),
			slog.String("owner_id", input.OperatorID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("save owned workbook list: %w", err)
	}

	now := time.Now()
	output, err := workbookservice.NewCreateWorkbookOutput(workbookID, input.SpaceID, input.OperatorID, input.OrganizationID, input.Title, input.Description, input.Visibility, now, now)
	if err != nil {
		return nil, fmt.Errorf("create workbook output: %w", err)
	}
	return output, nil
}

// grantWorkbookPolicies grants workbook-scoped permissions via RBAC.
func (c *CreateWorkbookCommand) grantWorkbookPolicies(ctx context.Context, organizationID, operatorID, workbookID string) error {
	actions := []domain.Action{
		domain.ActionViewWorkbook(),
		domain.ActionCreateQuestion(),
		domain.ActionUpdateQuestion(),
		domain.ActionDeleteQuestion(),
	}
	resource := domain.ResourceWorkbook(workbookID)
	for _, action := range actions {
		if err := c.policyAdder.AddPolicyForUser(ctx, organizationID, operatorID, action, resource, domain.EffectAllow); err != nil {
			return fmt.Errorf("add %s policy for workbook %s: %w", action.Value(), workbookID, err)
		}
	}
	return nil
}
