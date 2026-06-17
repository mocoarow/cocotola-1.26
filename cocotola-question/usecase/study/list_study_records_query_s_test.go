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

func newListStudyRecordsInput(t *testing.T) *studyservice.ListStudyRecordsInput {
	t.Helper()
	input, err := studyservice.NewListStudyRecordsInput(studyservice.ListStudyRecordsInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		WorkbookID:     fixtureWorkbookID,
	})
	require.NoError(t, err)
	return input
}

func Test_ListStudyRecordsQuery_shouldReturnRecords_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	now := fixtureClock
	records := []domainstudy.Record{
		*domainstudy.ReconstructRecord(fixtureWorkbookID, "q-1", 2, now, now.Add(24*time.Hour), 3, 1),
	}
	finder := newMockstudyRecordFinder(t)
	finder.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(records, nil)

	q := studyusecase.NewListStudyRecordsQuery(finder, workbookRepo, authChecker)

	// when
	output, err := q.ListStudyRecords(ctx, newListStudyRecordsInput(t))

	// then
	require.NoError(t, err)
	require.Len(t, output.Records, 1)
	assert.Equal(t, "q-1", output.Records[0].QuestionID)
	assert.Equal(t, 2, output.Records[0].ConsecutiveCorrect)
	assert.Equal(t, 3, output.Records[0].TotalCorrect)
	assert.Equal(t, 1, output.Records[0].TotalIncorrect)
}

func Test_ListStudyRecordsQuery_shouldSkipAuthCheck_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixturePublicWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)

	finder := newMockstudyRecordFinder(t)
	finder.On("FindByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return([]domainstudy.Record{}, nil)

	q := studyusecase.NewListStudyRecordsQuery(finder, workbookRepo, authChecker)

	// when
	output, err := q.ListStudyRecords(ctx, newListStudyRecordsInput(t))

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Records)
	authChecker.AssertNotCalled(t, "IsAllowed")
}

func Test_ListStudyRecordsQuery_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, nil)

	q := studyusecase.NewListStudyRecordsQuery(nil, workbookRepo, authChecker)

	// when
	_, err = q.ListStudyRecords(ctx, newListStudyRecordsInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ListStudyRecordsQuery_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	q := studyusecase.NewListStudyRecordsQuery(nil, workbookRepo, nil)

	// when
	_, err := q.ListStudyRecords(ctx, newListStudyRecordsInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}
