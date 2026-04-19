package workbook

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// DeleteWorkbookCommand handles workbook deletion.
type DeleteWorkbookCommand struct {
	workbookFinder  workbookFinder
	workbookDeleter workbookDeleter
	ownedListFinder ownedWorkbookListFinder
	ownedListSaver  ownedWorkbookListSaver
	authChecker     authorizationChecker
}

// NewDeleteWorkbookCommand returns a new DeleteWorkbookCommand.
func NewDeleteWorkbookCommand(
	workbookFinder workbookFinder,
	workbookDeleter workbookDeleter,
	ownedListFinder ownedWorkbookListFinder,
	ownedListSaver ownedWorkbookListSaver,
	authChecker authorizationChecker,
) *DeleteWorkbookCommand {
	return &DeleteWorkbookCommand{
		workbookFinder:  workbookFinder,
		workbookDeleter: workbookDeleter,
		ownedListFinder: ownedListFinder,
		ownedListSaver:  ownedListSaver,
		authChecker:     authChecker,
	}
}

// DeleteWorkbook deletes a workbook.
func (c *DeleteWorkbookCommand) DeleteWorkbook(ctx context.Context, input *workbookservice.DeleteWorkbookInput) error {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionDeleteWorkbook(), domain.ResourceAny())
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

	// Remove from owned workbook list.
	// NOTE: eventual consistency -- if Save fails, the workbook is deleted but remains in the owned list.
	ownedList, err := c.ownedListFinder.FindByOwnerID(ctx, input.OperatorID)
	if err != nil {
		return fmt.Errorf("find owned workbook list: %w", err)
	}
	ownedList.Remove(input.WorkbookID)
	if err := c.ownedListSaver.Save(ctx, ownedList); err != nil {
		slog.ErrorContext(ctx, "owned list save failed after workbook deletion",
			slog.String("workbook_id", input.WorkbookID),
			slog.String("owner_id", input.OperatorID),
			slog.Any("error", err),
		)
		return fmt.Errorf("save owned workbook list: %w", err)
	}

	return nil
}
