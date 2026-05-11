package study_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
)

func newDeleteStudyHistoryInput(t *testing.T) *studyservice.DeleteStudyHistoryInput {
	t.Helper()
	input, err := studyservice.NewDeleteStudyHistoryInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID)
	require.NoError(t, err)
	return input
}

func Test_DeleteStudyHistoryCommand_shouldDelete_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	deleter := newMockstudyRecordDeleter(t)
	deleter.On("DeleteByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(nil)

	cmd := studyusecase.NewDeleteStudyHistoryCommand(deleter, workbookRepo, authChecker)

	// when
	err = cmd.DeleteStudyHistory(ctx, newDeleteStudyHistoryInput(t))

	// then
	require.NoError(t, err)
}

func Test_DeleteStudyHistoryCommand_shouldSkipAuthCheck_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixturePublicWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)

	deleter := newMockstudyRecordDeleter(t)
	deleter.On("DeleteByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(nil)

	cmd := studyusecase.NewDeleteStudyHistoryCommand(deleter, workbookRepo, authChecker)

	// when
	err := cmd.DeleteStudyHistory(ctx, newDeleteStudyHistoryInput(t))

	// then
	require.NoError(t, err)
	authChecker.AssertNotCalled(t, "IsAllowed")
}

func Test_DeleteStudyHistoryCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, nil)

	deleter := newMockstudyRecordDeleter(t)

	cmd := studyusecase.NewDeleteStudyHistoryCommand(deleter, workbookRepo, authChecker)

	// when
	err = cmd.DeleteStudyHistory(ctx, newDeleteStudyHistoryInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
	deleter.AssertNotCalled(t, "DeleteByWorkbookID")
}

func Test_DeleteStudyHistoryCommand_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	deleter := newMockstudyRecordDeleter(t)

	cmd := studyusecase.NewDeleteStudyHistoryCommand(deleter, workbookRepo, nil)

	// when
	err := cmd.DeleteStudyHistory(ctx, newDeleteStudyHistoryInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
	deleter.AssertNotCalled(t, "DeleteByWorkbookID")
}

func Test_DeleteStudyHistoryCommand_shouldReturnError_whenDeleterFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	repoErr := errors.New("firestore unavailable")
	deleter := newMockstudyRecordDeleter(t)
	deleter.On("DeleteByWorkbookID", mock.Anything, fixtureOperatorID, fixtureWorkbookID).Return(repoErr)

	cmd := studyusecase.NewDeleteStudyHistoryCommand(deleter, workbookRepo, authChecker)

	// when
	err = cmd.DeleteStudyHistory(ctx, newDeleteStudyHistoryInput(t))

	// then
	require.ErrorIs(t, err, repoErr)
}
