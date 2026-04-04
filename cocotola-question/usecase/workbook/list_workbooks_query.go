package workbook

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// ListWorkbooksQuery handles listing workbooks in a space.
type ListWorkbooksQuery struct {
	workbookRepo workbookFinder
	authChecker  authorizationChecker
}

// NewListWorkbooksQuery returns a new ListWorkbooksQuery.
func NewListWorkbooksQuery(workbookRepo workbookFinder, authChecker authorizationChecker) *ListWorkbooksQuery {
	return &ListWorkbooksQuery{
		workbookRepo: workbookRepo,
		authChecker:  authChecker,
	}
}

// ListWorkbooks returns all workbooks in a space.
func (q *ListWorkbooksQuery) ListWorkbooks(ctx context.Context, input *workbookservice.ListWorkbooksInput) (*workbookservice.ListWorkbooksOutput, error) {
	allowed, err := q.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionViewWorkbook(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	workbooks, err := q.workbookRepo.FindBySpaceID(ctx, input.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("find workbooks by space: %w", err)
	}

	items := make([]workbookservice.Item, len(workbooks))
	for i, wb := range workbooks {
		items[i] = workbookservice.Item{
			WorkbookID:     wb.ID(),
			SpaceID:        wb.SpaceID(),
			OwnerID:        wb.OwnerID(),
			OrganizationID: wb.OrganizationID(),
			Title:          wb.Title(),
			Description:    wb.Description(),
			Visibility:     wb.Visibility().Value(),
			CreatedAt:      wb.CreatedAt(),
			UpdatedAt:      wb.UpdatedAt(),
		}
	}

	return &workbookservice.ListWorkbooksOutput{Workbooks: items}, nil
}
