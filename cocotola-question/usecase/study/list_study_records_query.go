package study

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// ListStudyRecordsQuery returns the operator's spaced-repetition records for a
// workbook. Authorization mirrors GetStudyQuestions / DeleteStudyHistory.
type ListStudyRecordsQuery struct {
	studyRecordFinder studyRecordFinder
	workbookRepo      workbookFinder
	authChecker       authorizationChecker
}

// NewListStudyRecordsQuery returns a new ListStudyRecordsQuery.
func NewListStudyRecordsQuery(
	studyRecordFinder studyRecordFinder,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
) *ListStudyRecordsQuery {
	return &ListStudyRecordsQuery{
		studyRecordFinder: studyRecordFinder,
		workbookRepo:      workbookRepo,
		authChecker:       authChecker,
	}
}

// ListStudyRecords returns every study record the operator has for the workbook.
func (q *ListStudyRecordsQuery) ListStudyRecords(ctx context.Context, input *studyservice.ListStudyRecordsInput) (*studyservice.ListStudyRecordsOutput, error) {
	wb, err := q.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if !wb.Visibility().IsPublic() {
		resource, err := domain.ResourceWorkbook(wb.ID())
		if err != nil {
			return nil, fmt.Errorf("resource workbook: %w", err)
		}
		allowed, err := q.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionStudyWorkbook(), resource)
		if err != nil {
			return nil, fmt.Errorf("authorization check: %w", err)
		}
		if !allowed {
			return nil, domain.ErrForbidden
		}
	}

	records, err := q.studyRecordFinder.FindByWorkbookID(ctx, input.OperatorID, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find study records: %w", err)
	}

	items := make([]studyservice.RecordItem, 0, len(records))
	for _, r := range records {
		items = append(items, studyservice.RecordItem{
			WorkbookID:         r.WorkbookID(),
			QuestionID:         r.QuestionID(),
			ConsecutiveCorrect: r.ConsecutiveCorrect(),
			LastAnsweredAt:     r.LastAnsweredAt(),
			NextDueAt:          r.NextDueAt(),
			TotalCorrect:       r.TotalCorrect(),
			TotalIncorrect:     r.TotalIncorrect(),
		})
	}

	return &studyservice.ListStudyRecordsOutput{Records: items}, nil
}
