package study

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// DeleteStudyHistoryCommand clears the operator's spaced-repetition history for
// a workbook. Questions themselves are kept; only the per-user study records
// are removed, so authorization mirrors GetStudyQuestions (public workbooks are
// open; private workbooks require explicit study permission).
type DeleteStudyHistoryCommand struct {
	studyRecordDeleter studyRecordDeleter
	workbookRepo       workbookFinder
	authChecker        authorizationChecker
}

// NewDeleteStudyHistoryCommand returns a new DeleteStudyHistoryCommand.
func NewDeleteStudyHistoryCommand(
	studyRecordDeleter studyRecordDeleter,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
) *DeleteStudyHistoryCommand {
	return &DeleteStudyHistoryCommand{
		studyRecordDeleter: studyRecordDeleter,
		workbookRepo:       workbookRepo,
		authChecker:        authChecker,
	}
}

// DeleteStudyHistory clears all study records the operator has for the workbook.
// Idempotent: returns nil even when no records exist.
func (c *DeleteStudyHistoryCommand) DeleteStudyHistory(ctx context.Context, input *studyservice.DeleteStudyHistoryInput) error {
	wb, err := c.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return fmt.Errorf("find workbook: %w", err)
	}

	if !wb.Visibility().IsPublic() {
		resource, err := domain.ResourceWorkbook(wb.ID())
		if err != nil {
			return fmt.Errorf("resource workbook: %w", err)
		}
		allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionStudyWorkbook(), resource)
		if err != nil {
			return fmt.Errorf("authorization check: %w", err)
		}
		if !allowed {
			return domain.ErrForbidden
		}
	}

	if err := c.studyRecordDeleter.DeleteByWorkbookID(ctx, input.OperatorID, input.WorkbookID); err != nil {
		return fmt.Errorf("delete study records: %w", err)
	}

	return nil
}
