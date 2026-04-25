package study

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// GetStudyQuestionsQuery handles study question retrieval.
type GetStudyQuestionsQuery struct {
	studyRecordRepo studyRecordFinder
	activeListRepo  activeQuestionListFinder
	questionRepo    questionBatchFinder
	workbookRepo    workbookFinder
	authChecker     authorizationChecker
	config          UsecaseConfig
}

// NewGetStudyQuestionsQuery returns a new GetStudyQuestionsQuery.
func NewGetStudyQuestionsQuery(
	studyRecordRepo studyRecordFinder,
	activeListRepo activeQuestionListFinder,
	questionRepo questionBatchFinder,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
	config UsecaseConfig,
) *GetStudyQuestionsQuery {
	return &GetStudyQuestionsQuery{
		studyRecordRepo: studyRecordRepo,
		activeListRepo:  activeListRepo,
		questionRepo:    questionRepo,
		workbookRepo:    workbookRepo,
		authChecker:     authChecker,
		config:          config,
	}
}

// GetStudyQuestions returns questions that are due for study.
func (q *GetStudyQuestionsQuery) GetStudyQuestions(ctx context.Context, input *studyservice.GetStudyQuestionsInput) (*studyservice.GetStudyQuestionsOutput, error) {
	wb, err := q.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

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

	activeList, err := q.activeListRepo.FindByWorkbookID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find active question list: %w", err)
	}

	studyRecords, err := q.studyRecordRepo.FindByWorkbookID(ctx, input.OperatorID, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find study records: %w", err)
	}

	dueIDs, newIDs, reviewCount, newCount := q.classifyQuestionIDs(activeList.Entries(), studyRecords)

	q.config.Shuffle(len(newIDs), func(i, j int) {
		newIDs[i], newIDs[j] = newIDs[j], newIDs[i]
	})

	selectedIDs := mixIDs(dueIDs, newIDs, input.Limit)
	totalDue := len(dueIDs) + len(newIDs)

	items, err := q.fetchQuestionItems(ctx, input.WorkbookID, selectedIDs)
	if err != nil {
		return nil, err
	}

	return &studyservice.GetStudyQuestionsOutput{
		Questions:   items,
		TotalDue:    totalDue,
		NewCount:    newCount,
		ReviewCount: reviewCount,
	}, nil
}

func (q *GetStudyQuestionsQuery) classifyQuestionIDs(activeIDs []string, studyRecords []domainstudy.Record) (dueIDs, newIDs []string, reviewCount, newCount int) {
	recordMap := make(map[string]int, len(studyRecords))
	for i, r := range studyRecords {
		recordMap[r.QuestionID()] = i
	}

	now := q.config.Now()
	for _, qID := range activeIDs {
		idx, hasRecord := recordMap[qID]
		if !hasRecord {
			newCount++
			newIDs = append(newIDs, qID)
		} else if !studyRecords[idx].NextDueAt().After(now) {
			reviewCount++
			dueIDs = append(dueIDs, qID)
		}
	}
	return
}

func (q *GetStudyQuestionsQuery) fetchQuestionItems(ctx context.Context, workbookID string, selectedIDs []string) ([]studyservice.QuestionItem, error) {
	if len(selectedIDs) == 0 {
		return nil, nil
	}

	questions, err := q.questionRepo.FindByIDs(ctx, workbookID, selectedIDs)
	if err != nil {
		return nil, fmt.Errorf("find questions by ids: %w", err)
	}

	items := make([]studyservice.QuestionItem, 0, len(questions))
	for _, question := range questions {
		items = append(items, studyservice.QuestionItem{
			QuestionID:   question.ID(),
			QuestionType: question.QuestionType().Value(),
			Content:      question.Content(),
			Tags:         question.Tags(),
			OrderIndex:   question.OrderIndex(),
		})
	}
	return items, nil
}

const reviewRatioNumerator = 9
const reviewRatioDenominator = 10

// mixIDs selects IDs with 90% review and 10% new ratio.
// If one pool has fewer than its allocated slots, the surplus is filled from the other.
func mixIDs(review, unseen []string, limit int) []string {
	reviewSlots := limit * reviewRatioNumerator / reviewRatioDenominator
	newSlots := limit - reviewSlots

	// Take from each pool, capped by availability
	takeReview := min(reviewSlots, len(review))
	takeNew := min(newSlots, len(unseen))

	// Fill surplus from the other pool
	remaining := limit - takeReview - takeNew
	if remaining > 0 {
		extraReview := min(remaining, len(review)-takeReview)
		takeReview += extraReview
		remaining -= extraReview
	}
	if remaining > 0 {
		extraNew := min(remaining, len(unseen)-takeNew)
		takeNew += extraNew
	}

	result := make([]string, 0, takeReview+takeNew)
	result = append(result, review[:takeReview]...)
	result = append(result, unseen[:takeNew]...)
	return result
}
