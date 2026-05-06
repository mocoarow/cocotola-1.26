package study

import (
	"context"
	"fmt"
	"time"

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

	// Public workbooks are studyable by all users; study records are scoped per user.
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

	dueIDs, newIDs, reviewCount, newCount := classifyQuestionIDs(activeList.Entries(), studyRecords, q.config.Now(), input.Practice)

	q.config.Shuffle(len(newIDs), func(i, j int) {
		newIDs[i], newIDs[j] = newIDs[j], newIDs[i]
	})

	selectedIDs := mixIDs(dueIDs, newIDs, input.Limit)
	totalDue := len(dueIDs) + len(newIDs)

	items, err := q.fetchQuestionItems(ctx, input.WorkbookID, selectedIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch question items: %w", err)
	}

	return &studyservice.GetStudyQuestionsOutput{
		Questions:   items,
		TotalDue:    totalDue,
		NewCount:    newCount,
		ReviewCount: reviewCount,
	}, nil
}

// classifyQuestionIDs splits the active question pool into "new" (no record)
// and "due" (record exists and is past its NextDueAt) buckets. In practice
// mode the schedule is ignored: every active question with a record is treated
// as due so the user can keep solving even after the day's queue is exhausted,
// regardless of whether the question was previously answered correctly or
// incorrectly. Shared between GetStudyQuestionsQuery and GetStudySummaryQuery.
func classifyQuestionIDs(activeIDs []string, studyRecords []domainstudy.Record, now time.Time, practice bool) (dueIDs, newIDs []string, reviewCount, newCount int) {
	recordMap := make(map[string]int, len(studyRecords))
	for i, r := range studyRecords {
		recordMap[r.QuestionID()] = i
	}

	for _, qID := range activeIDs {
		idx, hasRecord := recordMap[qID]
		switch {
		case !hasRecord:
			newCount++
			newIDs = append(newIDs, qID)
		case practice || !studyRecords[idx].NextDueAt().After(now):
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

// ReviewRatioNumerator and ReviewRatioDenominator define the fixed
// review/new mix used by mixIDs when filling a study session. Exposed so
// callers (e.g. the summary endpoint) can advertise the ratio to clients
// without duplicating the constants.
const (
	ReviewRatioNumerator   = 9
	ReviewRatioDenominator = 10
)

// mixIDs selects IDs with 90% review and 10% new ratio.
// If one pool has fewer than its allocated slots, the surplus is filled from the other.
func mixIDs(review, unseen []string, limit int) []string {
	reviewSlots := limit * ReviewRatioNumerator / ReviewRatioDenominator
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
