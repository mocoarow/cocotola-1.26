package question_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

const fixtureOtherQuestionID = "q-other"

func fixtureVersionedActiveQuestionList(t *testing.T, version int, questionIDs ...string) *domain.ActiveQuestionList {
	t.Helper()
	list := fixtureActiveQuestionList(t, questionIDs...)
	list.SetVersion(version)
	return list
}

func allowWorkbookAction(t *testing.T, action domain.Action) *mockauthorizationChecker {
	t.Helper()
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, action, wbResource).Return(true, nil)
	return authChecker
}

func acceptingQuestionDeleter(t *testing.T) *mockquestionDeleter {
	t.Helper()
	questionDeleter := newMockquestionDeleter(t)
	questionDeleter.On("Delete", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(nil)
	return questionDeleter
}

func Test_DeleteQuestionCommand_shouldDeleteQuestion_whenActiveListSaveConflictsOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

	cmd := questionusecase.NewDeleteQuestionCommand(acceptingQuestionDeleter(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	err := cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	require.NoError(t, err)
}

func Test_DeleteQuestionCommand_shouldReturnConcurrentModification_whenActiveListKeepsConflicting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	cmd := questionusecase.NewDeleteQuestionCommand(acceptingQuestionDeleter(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	err := cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_DeleteQuestionCommand_shouldStopRetrying_whenActiveListSaveAttemptsAreExhausted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	cmd := questionusecase.NewDeleteQuestionCommand(acceptingQuestionDeleter(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	_ = cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	activeListSaver.AssertNumberOfCalls(t, "Save", questionusecase.MaxActiveListSaveAttempts)
}

func Test_DeleteQuestionCommand_shouldNotRetry_whenActiveListSaveFailsWithOtherError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	saveErr := errors.New("firestore unavailable")
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(saveErr)

	cmd := questionusecase.NewDeleteQuestionCommand(acceptingQuestionDeleter(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	_ = cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	activeListSaver.AssertNumberOfCalls(t, "Save", 1)
}

func Test_DeleteQuestionCommand_shouldSaveRereadActiveList_whenActiveListSaveConflicts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil).Once()
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 2, fixtureQuestionID, fixtureOtherQuestionID), nil).Once()
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeListSaver.On("Save", mock.Anything, mock.MatchedBy(func(l *domain.ActiveQuestionList) bool {
		return l.Version() == 2 && l.Size() == 1 && l.Contains(fixtureOtherQuestionID)
	})).Return(nil).Once()

	cmd := questionusecase.NewDeleteQuestionCommand(acceptingQuestionDeleter(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	err := cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	require.NoError(t, err)
}
