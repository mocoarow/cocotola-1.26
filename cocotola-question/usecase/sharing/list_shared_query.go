package sharing

import (
	"context"
	"fmt"

	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

// ListSharedQuery handles listing shared workbook references.
type ListSharedQuery struct {
	referenceRepo referenceFinder
	workbookRepo  workbookBatchFinder
}

// NewListSharedQuery returns a new ListSharedQuery.
func NewListSharedQuery(referenceRepo referenceFinder, workbookRepo workbookBatchFinder) *ListSharedQuery {
	return &ListSharedQuery{
		referenceRepo: referenceRepo,
		workbookRepo:  workbookRepo,
	}
}

// ListShared returns the operator's workbook references whose workbooks are
// public in the operator's organization.
func (q *ListSharedQuery) ListShared(ctx context.Context, input *referenceservice.ListSharedInput) (*referenceservice.ListSharedOutput, error) {
	refs, err := q.referenceRepo.FindByUserID(ctx, input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("find references: %w", err)
	}
	if len(refs) == 0 {
		return &referenceservice.ListSharedOutput{References: []referenceservice.SharedItem{}}, nil
	}

	workbookIDs := make([]string, len(refs))
	for i, ref := range refs {
		workbookIDs[i] = ref.WorkbookID()
	}
	workbooks, err := q.workbookRepo.FindByIDs(ctx, workbookIDs)
	if err != nil {
		return nil, fmt.Errorf("find workbooks by ids: %w", err)
	}
	workbookByID := make(map[string]domainworkbook.Workbook, len(workbooks))
	for _, wb := range workbooks {
		workbookByID[wb.ID()] = wb
	}

	items := make([]referenceservice.SharedItem, 0, len(refs))
	for _, ref := range refs {
		wb, ok := workbookByID[ref.WorkbookID()]
		if !ok || !isListable(&wb, input) {
			continue
		}
		items = append(items, referenceservice.SharedItem{
			ReferenceID: ref.ID(),
			WorkbookID:  ref.WorkbookID(),
			Title:       wb.Title(),
			Description: wb.Description(),
			Language:    wb.Language().Value(),
			AddedAt:     ref.AddedAt(),
		})
	}

	return &referenceservice.ListSharedOutput{References: items}, nil
}

func isListable(wb *domainworkbook.Workbook, input *referenceservice.ListSharedInput) bool {
	return wb.Visibility().IsPublic() && wb.OrganizationID() == input.OrganizationID
}
