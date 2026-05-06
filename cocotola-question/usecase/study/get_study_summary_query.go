package study

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// GetStudySummaryQuery returns the available review/new question counts for a
// workbook without loading the question payloads themselves. This powers the
// pre-study size picker so clients can render an informed dialog without
// downloading every question.
type GetStudySummaryQuery struct {
	studyRecordRepo studyRecordFinder
	activeListRepo  activeQuestionListFinder
	workbookRepo    workbookFinder
	authChecker     authorizationChecker
	config          UsecaseConfig
}

// NewGetStudySummaryQuery returns a new GetStudySummaryQuery.
func NewGetStudySummaryQuery(
	studyRecordRepo studyRecordFinder,
	activeListRepo activeQuestionListFinder,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
	config UsecaseConfig,
) *GetStudySummaryQuery {
	return &GetStudySummaryQuery{
		studyRecordRepo: studyRecordRepo,
		activeListRepo:  activeListRepo,
		workbookRepo:    workbookRepo,
		authChecker:     authChecker,
		config:          config,
	}
}

// GetStudySummary returns the count of due/new questions for the workbook.
// Authorization mirrors GetStudyQuestions: public workbooks are studyable by
// all users, private workbooks require explicit study permission.
func (q *GetStudySummaryQuery) GetStudySummary(ctx context.Context, input *studyservice.GetStudySummaryInput) (*studyservice.GetStudySummaryOutput, error) {
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

	activeList, err := q.activeListRepo.FindByWorkbookID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find active question list: %w", err)
	}

	studyRecords, err := q.studyRecordRepo.FindByWorkbookID(ctx, input.OperatorID, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find study records: %w", err)
	}

	_, _, reviewCount, newCount := classifyQuestionIDs(activeList.Entries(), studyRecords, q.config.Now(), input.Practice)

	return &studyservice.GetStudySummaryOutput{
		NewCount:               newCount,
		ReviewCount:            reviewCount,
		TotalDue:               reviewCount + newCount,
		ReviewRatioNumerator:   ReviewRatioNumerator,
		ReviewRatioDenominator: ReviewRatioDenominator,
	}, nil
}
