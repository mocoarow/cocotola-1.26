package workbook

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// GetWorkbookQuery handles workbook retrieval.
type GetWorkbookQuery struct {
	workbookRepo workbookFinder
	authChecker  authorizationChecker
}

// NewGetWorkbookQuery returns a new GetWorkbookQuery.
func NewGetWorkbookQuery(workbookRepo workbookFinder, authChecker authorizationChecker) *GetWorkbookQuery {
	return &GetWorkbookQuery{
		workbookRepo: workbookRepo,
		authChecker:  authChecker,
	}
}

// GetWorkbook retrieves a workbook by ID.
func (q *GetWorkbookQuery) GetWorkbook(ctx context.Context, input *workbookservice.GetWorkbookInput) (*workbookservice.GetWorkbookOutput, error) {
	wb, err := q.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	// Public workbooks are accessible to all
	if !wb.Visibility().IsPublic() {
		resource, err := domain.ResourceWorkbook(input.WorkbookID)
		if err != nil {
			return nil, fmt.Errorf("resource workbook: %w", err)
		}
		allowed, err := q.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionViewWorkbook(), resource)
		if err != nil {
			return nil, fmt.Errorf("authorization check: %w", err)
		}
		if !allowed {
			return nil, domain.ErrForbidden
		}
	}

	return &workbookservice.GetWorkbookOutput{
		Item: workbookservice.Item{
			WorkbookID:     wb.ID(),
			SpaceID:        wb.SpaceID(),
			OwnerID:        wb.OwnerID(),
			OrganizationID: wb.OrganizationID(),
			Title:          wb.Title(),
			Description:    wb.Description(),
			Visibility:     wb.Visibility().Value(),
			CreatedAt:      wb.CreatedAt(),
			UpdatedAt:      wb.UpdatedAt(),
		},
	}, nil
}
