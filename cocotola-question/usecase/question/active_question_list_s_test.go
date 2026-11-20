package question_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
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

func acceptingQuestionSaver(t *testing.T) *mockquestionSaver {
	t.Helper()
	questionSaver := newMockquestionSaver(t)
	questionSaver.On("Save", mock.Anything, mock.Anything).Return(nil)
	return questionSaver
}

func Test_AddQuestionCommand_shouldAddQuestion_whenActiveListSaveConflictsOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

	cmd := questionusecase.NewAddQuestionCommand(acceptingQuestionSaver(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, err := cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then
	require.NoError(t, err)
}

func Test_AddQuestionCommand_shouldSaveRereadActiveList_whenActiveListSaveConflicts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1), nil).Once()
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 2, fixtureOtherQuestionID), nil).Once()
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeListSaver.On("Save", mock.Anything, mock.MatchedBy(func(l *domain.ActiveQuestionList) bool {
		return l.Version() == 2 && l.Size() == 2 && l.Contains(fixtureOtherQuestionID)
	})).Return(nil).Once()

	cmd := questionusecase.NewAddQuestionCommand(acceptingQuestionSaver(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, err := cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then
	require.NoError(t, err)
}

func Test_AddQuestionCommand_shouldReturnConcurrentModification_whenActiveListKeepsConflicting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	cmd := questionusecase.NewAddQuestionCommand(acceptingQuestionSaver(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, err := cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_AddQuestionCommand_shouldStopRetrying_whenActiveListSaveAttemptsAreExhausted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	cmd := questionusecase.NewAddQuestionCommand(acceptingQuestionSaver(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, _ = cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then
	activeListSaver.AssertNumberOfCalls(t, "Save", questionusecase.MaxActiveListSaveAttempts)
}

func Test_AddQuestionCommand_shouldNotRetry_whenActiveListSaveFailsWithOtherError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	saveErr := errors.New("firestore unavailable")
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(saveErr)

	cmd := questionusecase.NewAddQuestionCommand(acceptingQuestionSaver(t), activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, _ = cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then
	activeListSaver.AssertNumberOfCalls(t, "Save", 1)
}

func Test_DeleteQuestionCommand_shouldDeleteQuestion_whenActiveListSaveConflictsOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	questionDeleter := newMockquestionDeleter(t)
	questionDeleter.On("Delete", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(nil)
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

	cmd := questionusecase.NewDeleteQuestionCommand(questionDeleter, activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	err := cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	require.NoError(t, err)
}

func Test_DeleteQuestionCommand_shouldReturnConcurrentModification_whenActiveListKeepsConflicting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	questionDeleter := newMockquestionDeleter(t)
	questionDeleter.On("Delete", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(nil)
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil)
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	cmd := questionusecase.NewDeleteQuestionCommand(questionDeleter, activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	err := cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_DeleteQuestionCommand_shouldSaveRereadActiveList_whenActiveListSaveConflicts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	questionDeleter := newMockquestionDeleter(t)
	questionDeleter.On("Delete", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(nil)
	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 1, fixtureQuestionID), nil).Once()
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureVersionedActiveQuestionList(t, 2, fixtureQuestionID, fixtureOtherQuestionID), nil).Once()
	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeListSaver.On("Save", mock.Anything, mock.MatchedBy(func(l *domain.ActiveQuestionList) bool {
		return l.Version() == 2 && l.Size() == 1 && l.Contains(fixtureOtherQuestionID)
	})).Return(nil).Once()

	cmd := questionusecase.NewDeleteQuestionCommand(questionDeleter, activeListFinder, activeListSaver, allowWorkbookAction(t, domain.ActionDeleteQuestion()))

	// when
	err := cmd.DeleteQuestion(ctx, newDeleteQuestionInput(t))

	// then
	require.NoError(t, err)
}
