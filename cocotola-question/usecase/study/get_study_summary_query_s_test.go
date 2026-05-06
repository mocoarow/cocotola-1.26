package study_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
)

func newGetStudySummaryInput(t *testing.T) *studyservice.GetStudySummaryInput {
	t.Helper()
	input, err := studyservice.NewGetStudySummaryInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, false)
	require.NoError(t, err)
	return input
}

func newGetStudySummaryInputForPractice(t *testing.T) *studyservice.GetStudySummaryInput {
	t.Helper()
	input, err := studyservice.NewGetStudySummaryInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, true)
	require.NoError(t, err)
	return input
}

func Test_GetStudySummaryQuery_shouldReportCounts_whenMixedRecords(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: three active questions — one with a past-due record, one with a
	// future-due record, one without any record at all.
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, "q-1", "q-2", "q-3"), nil)

	pastDue := fixtureClock.Add(-24 * time.Hour)
	futureDue := fixtureClock.Add(24 * time.Hour)
	records := []domainstudy.Record{
		*domainstudy.ReconstructRecord(fixtureWorkbookID, "q-1", 1, pastDue, pastDue, 1, 0),
		*domainstudy.ReconstructRecord(fixtureWorkbookID, "q-2", 1, fixtureClock, futureDue, 1, 0),
	}
	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	query := studyusecase.NewGetStudySummaryQuery(studyRecordRepo, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudySummaryInput(t)

	// when
	output, err := query.GetStudySummary(ctx, input)

	// then: q-1 is review (past due), q-3 is new, q-2 is suppressed (not due).
	require.NoError(t, err)
	assert.Equal(t, 1, output.ReviewCount)
	assert.Equal(t, 1, output.NewCount)
	assert.Equal(t, 2, output.TotalDue)
	assert.Equal(t, studyusecase.ReviewRatioNumerator, output.ReviewRatioNumerator)
	assert.Equal(t, studyusecase.ReviewRatioDenominator, output.ReviewRatioDenominator)
}

func Test_GetStudySummaryQuery_shouldIncludeNotYetDue_whenPracticeMode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a not-yet-due record should still be counted as review in practice mode.
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

	query := studyusecase.NewGetStudySummaryQuery(studyRecordRepo, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudySummaryInputForPractice(t)

	// when
	output, err := query.GetStudySummary(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, output.ReviewCount)
	assert.Equal(t, 0, output.NewCount)
	assert.Equal(t, 1, output.TotalDue)
}

func Test_GetStudySummaryQuery_shouldReportZero_whenActiveListIsEmpty(t *testing.T) {
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

	query := studyusecase.NewGetStudySummaryQuery(studyRecordRepo, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudySummaryInput(t)

	// when
	output, err := query.GetStudySummary(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, output.ReviewCount)
	assert.Equal(t, 0, output.NewCount)
	assert.Equal(t, 0, output.TotalDue)
}

func Test_GetStudySummaryQuery_shouldSkipAuthCheck_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixturePublicWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	studyRecordRepo := newMockstudyRecordFinder(t)
	studyRecordRepo.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return([]domainstudy.Record{}, nil)

	query := studyusecase.NewGetStudySummaryQuery(studyRecordRepo, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newGetStudySummaryInput(t)

	// when
	output, err := query.GetStudySummary(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, output.NewCount)
	authChecker.AssertNotCalled(t, "IsAllowed")
}

func Test_GetStudySummaryQuery_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, nil)

	query := studyusecase.NewGetStudySummaryQuery(nil, nil, workbookRepo, authChecker, testConfig)
	input := newGetStudySummaryInput(t)

	// when
	_, err = query.GetStudySummary(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_GetStudySummaryQuery_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	query := studyusecase.NewGetStudySummaryQuery(nil, nil, workbookRepo, nil, testConfig)
	input := newGetStudySummaryInput(t)

	// when
	_, err := query.GetStudySummary(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}
