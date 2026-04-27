package sharing

import (
	"context"
	"fmt"

	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
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

// ListPublic returns all public workbooks for the organization that match
// the requested language.
func (q *ListPublicQuery) ListPublic(ctx context.Context, input *referenceservice.ListPublicInput) (*referenceservice.ListPublicOutput, error) {
	// Re-validate language via the domain value object so the Firestore
	// query never receives a value that slipped past service-layer
	// length checks (e.g. uppercase, digits).
	lang, err := domainworkbook.NewLanguage(input.Language)
	if err != nil {
		return nil, fmt.Errorf("new language: %w", err)
	}

	workbooks, err := q.workbookRepo.FindPublicByOrganizationIDAndLanguage(ctx, input.OrganizationID, lang.Value())
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
			Language:    wb.Language().Value(),
			CreatedAt:   wb.CreatedAt(),
		}
	}

	return &referenceservice.ListPublicOutput{Workbooks: items}, nil
}
