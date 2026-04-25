package study_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
)

func newGetStudyQuestionsInput(t *testing.T, limit int) *studyservice.GetStudyQuestionsInput {
	t.Helper()
	input, err := studyservice.NewGetStudyQuestionsInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, limit)
	require.NoError(t, err)
	return input
}

func Test_GetStudyQuestionsQuery_shouldReturnNewQuestions_whenNoStudyRecords(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	questionRepo := newMockquestionBatchFinder(t)
	questionRepo.On("FindByIDs", mock.Anything, fixtureWorkbookID, mock.Anything).Return(fixtureQuestions(), nil)

	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return([]domainstudy.Record{}, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 1)
	assert.Equal(t, fixtureQuestionID, output.Questions[0].QuestionID)
	assert.Equal(t, 1, output.TotalDue)
	assert.Equal(t, 1, output.NewCount)
	assert.Equal(t, 0, output.ReviewCount)
}

func Test_GetStudyQuestionsQuery_shouldReturnDueQuestions_whenRecordIsDue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	questionRepo := newMockquestionBatchFinder(t)
	questionRepo.On("FindByIDs", mock.Anything, fixtureWorkbookID, mock.Anything).Return(fixtureQuestions(), nil)

	pastDue := fixtureClock.Add(-24 * time.Hour)
	records := []domainstudy.Record{
		*domainstudy.ReconstructRecord(fixtureWorkbookID, fixtureQuestionID, 1, pastDue, pastDue, 1, 0),
	}
	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 1)
	assert.Equal(t, 1, output.ReviewCount)
	assert.Equal(t, 0, output.NewCount)
}

func Test_GetStudyQuestionsQuery_shouldReturnEmpty_whenNotDue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	futureDue := fixtureClock.Add(24 * time.Hour)
	records := []domainstudy.Record{
		*domainstudy.ReconstructRecord(fixtureWorkbookID, fixtureQuestionID, 1, fixtureClock, futureDue, 1, 0),
	}
	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, nil, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Questions)
	assert.Equal(t, 0, output.TotalDue)
}

func Test_GetStudyQuestionsQuery_shouldRespectLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	activeIDs := []string{"q-1", "q-2", "q-3"}

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, activeIDs...), nil)

	questions := []domainquestion.Question{
		*domainquestion.ReconstructQuestion("q-1", domainquestion.TypeWordFill(), `{"source":"a","target":"b","sourceLang":"en","targetLang":"ja","blanks":["a"]}`, nil, 0, now, now),
		*domainquestion.ReconstructQuestion("q-2", domainquestion.TypeWordFill(), `{"source":"c","target":"d","sourceLang":"en","targetLang":"ja","blanks":["c"]}`, nil, 1, now, now),
	}
	questionRepo := newMockquestionBatchFinder(t)
	questionRepo.On("FindByIDs", mock.Anything, fixtureWorkbookID, mock.Anything).Return(questions, nil)

	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return([]domainstudy.Record{}, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 2)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 2)
	assert.Equal(t, 3, output.TotalDue)
	assert.Equal(t, 3, output.NewCount)
}

func Test_GetStudyQuestionsQuery_shouldMix90PercentReviewAnd10PercentNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	pastDue := fixtureClock.Add(-24 * time.Hour)

	// 10 review questions + 10 new questions
	activeIDs := make([]string, 0, 20)
	allQuestions := make([]domainquestion.Question, 0, 20)
	var records []domainstudy.Record
	for i := range 20 {
		qID := "q-" + string(rune('a'+i))
		activeIDs = append(activeIDs, qID)
		allQuestions = append(allQuestions, *domainquestion.ReconstructQuestion(qID, domainquestion.TypeWordFill(), `{"source":"a","target":"b","sourceLang":"en","targetLang":"ja","blanks":["a"]}`, nil, i, now, now))
		if i < 10 {
			// First 10 have study records (due)
			records = append(records, *domainstudy.ReconstructRecord(fixtureWorkbookID, qID, 1, pastDue, pastDue, 1, 0))
		}
	}

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, activeIDs...), nil)

	questionRepo := newMockquestionBatchFinder(t)
	questionRepo.On("FindByIDs", mock.Anything, fixtureWorkbookID, mock.Anything).Return(allQuestions[:10], nil)

	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 10)
	assert.Equal(t, 10, output.ReviewCount)
	assert.Equal(t, 10, output.NewCount)
	assert.Equal(t, 20, output.TotalDue)
}

func Test_GetStudyQuestionsQuery_shouldFillWithNew_whenNotEnoughReview(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	pastDue := fixtureClock.Add(-24 * time.Hour)

	// 2 review questions + 10 new questions
	activeIDs := make([]string, 0, 12)
	allQuestions := make([]domainquestion.Question, 0, 12)
	var records []domainstudy.Record
	for i := range 12 {
		qID := "q-" + string(rune('a'+i))
		activeIDs = append(activeIDs, qID)
		allQuestions = append(allQuestions, *domainquestion.ReconstructQuestion(qID, domainquestion.TypeWordFill(), `{"source":"a","target":"b","sourceLang":"en","targetLang":"ja","blanks":["a"]}`, nil, i, now, now))
		if i < 2 {
			records = append(records, *domainstudy.ReconstructRecord(fixtureWorkbookID, qID, 1, pastDue, pastDue, 1, 0))
		}
	}

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, activeIDs...), nil)

	questionRepo := newMockquestionBatchFinder(t)
	questionRepo.On("FindByIDs", mock.Anything, fixtureWorkbookID, mock.Anything).Return(allQuestions[:10], nil)

	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 10)
	assert.Equal(t, 2, output.ReviewCount)
	assert.Equal(t, 10, output.NewCount)
}

func Test_GetStudyQuestionsQuery_shouldReturnEmpty_whenActiveListIsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t), nil)

	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return([]domainstudy.Record{}, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, nil, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Questions)
	assert.Equal(t, 0, output.TotalDue)
	assert.Equal(t, 0, output.NewCount)
	assert.Equal(t, 0, output.ReviewCount)
}

func Test_GetStudyQuestionsQuery_shouldReturnOneQuestion_whenLimitIsOne(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	pastDue := fixtureClock.Add(-24 * time.Hour)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, "q-1", "q-2"), nil)

	questions := []domainquestion.Question{
		*domainquestion.ReconstructQuestion("q-1", domainquestion.TypeWordFill(), `{"source":"a","target":"b","sourceLang":"en","targetLang":"ja","blanks":["a"]}`, nil, 0, now, now),
	}
	questionRepo := newMockquestionBatchFinder(t)
	questionRepo.On("FindByIDs", mock.Anything, fixtureWorkbookID, mock.Anything).Return(questions, nil)

	records := []domainstudy.Record{
		*domainstudy.ReconstructRecord(fixtureWorkbookID, "q-1", 1, pastDue, pastDue, 1, 0),
	}
	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 1)

	// when
	output, err := query.GetStudyQuestions(ctx, input)

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 1)
	assert.Equal(t, 2, output.TotalDue)
}

func Test_GetStudyQuestionsQuery_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, nil)

	query := studyusecase.NewGetStudyQuestionsQuery(nil, nil, nil, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	_, err = query.GetStudyQuestions(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_GetStudyQuestionsQuery_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	query := studyusecase.NewGetStudyQuestionsQuery(nil, nil, nil, workbookRepo, nil, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	_, err := query.GetStudyQuestions(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}

func Test_GetStudyQuestionsQuery_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authErr := errors.New("auth unavailable")
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, authErr)

	query := studyusecase.NewGetStudyQuestionsQuery(nil, nil, nil, workbookRepo, authChecker, testConfig)
	input := newGetStudyQuestionsInput(t, 10)

	// when
	_, err = query.GetStudyQuestions(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}
