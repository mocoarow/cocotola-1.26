package sharing

import (
	"context"
	"fmt"

	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

// ListSharedQuery handles listing shared workbook references.
type ListSharedQuery struct {
	referenceRepo referenceFinder
}

// NewListSharedQuery returns a new ListSharedQuery.
func NewListSharedQuery(referenceRepo referenceFinder) *ListSharedQuery {
	return &ListSharedQuery{
		referenceRepo: referenceRepo,
	}
}

// ListShared returns all shared workbook references for the operator.
func (q *ListSharedQuery) ListShared(ctx context.Context, input *referenceservice.ListSharedInput) (*referenceservice.ListSharedOutput, error) {
	refs, err := q.referenceRepo.FindByUserID(ctx, input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("find references: %w", err)
	}

	items := make([]referenceservice.SharedItem, len(refs))
	for i, ref := range refs {
		items[i] = referenceservice.SharedItem{
			ReferenceID: ref.ID(),
			WorkbookID:  ref.WorkbookID(),
			AddedAt:     ref.AddedAt(),
		}
	}

	return &referenceservice.ListSharedOutput{References: items}, nil
}
