package workbook_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
	workbookusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/workbook"
)

func newDeleteWorkbookInput(t *testing.T) *workbookservice.DeleteWorkbookInput {
	t.Helper()
	input, err := workbookservice.NewDeleteWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID)
	require.NoError(t, err)
	return input
}

func newFixtureWorkbook(ownerID string) *domainworkbook.Workbook {
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, fixtureSpaceID, ownerID, fixtureOrganizationID, "title", "desc", domainworkbook.VisibilityPrivate(), time.Now(), time.Now())
}

func Test_DeleteWorkbookCommand_shouldDeleteWorkbook_whenOwnerDeletes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteWorkbook(), domain.ResourceWorkbook(fixtureWorkbookID)).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newFixtureWorkbook(fixtureOperatorID), nil)

	wbDeleter := newMockworkbookDeleter(t)
	wbDeleter.On("Delete", mock.Anything, fixtureWorkbookID).Return(nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, []string{fixtureWorkbookID})
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	listSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	cmd := workbookusecase.NewDeleteWorkbookCommand(wbFinder, wbDeleter, listFinder, listSaver, authChecker)
	input := newDeleteWorkbookInput(t)

	// when
	err := cmd.DeleteWorkbook(ctx, input)

	// then
	require.NoError(t, err)
}

func Test_DeleteWorkbookCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteWorkbook(), domain.ResourceWorkbook(fixtureWorkbookID)).Return(false, nil)

	cmd := workbookusecase.NewDeleteWorkbookCommand(nil, nil, nil, nil, authChecker)
	input := newDeleteWorkbookInput(t)

	// when
	err := cmd.DeleteWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_DeleteWorkbookCommand_shouldReturnForbidden_whenNotOwner(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteWorkbook(), domain.ResourceWorkbook(fixtureWorkbookID)).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newFixtureWorkbook("other-user"), nil)

	cmd := workbookusecase.NewDeleteWorkbookCommand(wbFinder, nil, nil, nil, authChecker)
	input := newDeleteWorkbookInput(t)

	// when
	err := cmd.DeleteWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_DeleteWorkbookCommand_shouldReturnError_whenOwnedListSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteWorkbook(), domain.ResourceWorkbook(fixtureWorkbookID)).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newFixtureWorkbook(fixtureOperatorID), nil)

	wbDeleter := newMockworkbookDeleter(t)
	wbDeleter.On("Delete", mock.Anything, fixtureWorkbookID).Return(nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, []string{fixtureWorkbookID})
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	saveErr := errors.New("firestore unavailable")
	listSaver.On("Save", mock.Anything, mock.Anything).Return(saveErr)

	cmd := workbookusecase.NewDeleteWorkbookCommand(wbFinder, wbDeleter, listFinder, listSaver, authChecker)
	input := newDeleteWorkbookInput(t)

	// when
	err := cmd.DeleteWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, saveErr)
}
