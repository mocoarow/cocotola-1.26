package sharing

import (
	"context"
	"fmt"

	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

// ListPublicQuery handles listing public workbooks in an organization.
type ListPublicQuery struct {
	workbookRepo publicWorkbookFinder
}

// NewListPublicQuery returns a new ListPublicQuery.
func NewListPublicQuery(workbookRepo publicWorkbookFinder) *ListPublicQuery {
	return &ListPublicQuery{
		workbookRepo: workbookRepo,
	}
}

// ListPublic returns all public workbooks for the organization.
func (q *ListPublicQuery) ListPublic(ctx context.Context, input *referenceservice.ListPublicInput) (*referenceservice.ListPublicOutput, error) {
	workbooks, err := q.workbookRepo.FindPublicByOrganizationID(ctx, input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("find public workbooks: %w", err)
	}

	items := make([]referenceservice.PublicWorkbookItem, len(workbooks))
	for i, wb := range workbooks {
		items[i] = referenceservice.PublicWorkbookItem{
			WorkbookID:  wb.ID(),
			OwnerID:     wb.OwnerID(),
			Title:       wb.Title(),
			Description: wb.Description(),
			CreatedAt:   wb.CreatedAt(),
		}
	}

	return &referenceservice.ListPublicOutput{Workbooks: items}, nil
}
