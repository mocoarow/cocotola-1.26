package workbook

import (
	"context"
	"fmt"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// DeleteWorkbookCommand handles workbook deletion.
type DeleteWorkbookCommand struct {
	workbookFinder  workbookFinder
	workbookDeleter workbookDeleter
	authChecker     authorizationChecker
}

// NewDeleteWorkbookCommand returns a new DeleteWorkbookCommand.
func NewDeleteWorkbookCommand(workbookFinder workbookFinder, workbookDeleter workbookDeleter, authChecker authorizationChecker) *DeleteWorkbookCommand {
	return &DeleteWorkbookCommand{
		workbookFinder:  workbookFinder,
		workbookDeleter: workbookDeleter,
		authChecker:     authChecker,
	}
}

// DeleteWorkbook deletes a workbook.
func (c *DeleteWorkbookCommand) DeleteWorkbook(ctx context.Context, input *workbookservice.DeleteWorkbookInput) error {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domainrbac.ActionDeleteWorkbook(), domainrbac.ResourceAny())
	if err != nil {
		return fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return domain.ErrForbidden
	}

	wb, err := c.workbookFinder.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return fmt.Errorf("find workbook: %w", err)
	}

	if wb.OwnerID() != input.OperatorID {
		return domain.ErrForbidden
	}

	if err := c.workbookDeleter.Delete(ctx, input.WorkbookID); err != nil {
		return fmt.Errorf("delete workbook: %w", err)
	}

	return nil
}
