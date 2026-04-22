package question_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

func newDeleteQuestionInput(t *testing.T) *questionservice.DeleteQuestionInput {
	t.Helper()
	input, err := questionservice.NewDeleteQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, fixtureQuestionID)
	require.NoError(t, err)
	return input
}

func Test_DeleteQuestionCommand_shouldDeleteQuestion_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteQuestion(), wbResource).Return(true, nil)

	questionDeleter := newMockquestionDeleter(t)
	questionDeleter.On("Delete", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(nil)

	cmd := questionusecase.NewDeleteQuestionCommand(questionDeleter, authChecker)
	input := newDeleteQuestionInput(t)

	// when
	err = cmd.DeleteQuestion(ctx, input)

	// then
	require.NoError(t, err)
}

func Test_DeleteQuestionCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteQuestion(), wbResource).Return(false, nil)

	cmd := questionusecase.NewDeleteQuestionCommand(nil, authChecker)
	input := newDeleteQuestionInput(t)

	// when
	err = cmd.DeleteQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_DeleteQuestionCommand_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authErr := errors.New("auth unavailable")
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteQuestion(), wbResource).Return(false, authErr)

	cmd := questionusecase.NewDeleteQuestionCommand(nil, authChecker)
	input := newDeleteQuestionInput(t)

	// when
	err = cmd.DeleteQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}
