package question_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

const (
	fixtureOperatorID     = "user-1"
	fixtureOrganizationID = "org-1"
	fixtureWorkbookID     = "wb-1"
	fixtureQuestionID     = "q-1"
)

func newAddQuestionInput(t *testing.T) *questionservice.AddQuestionInput {
	t.Helper()
	input, err := questionservice.NewAddQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, "word_fill", "What is Go?", []string{"lang:en"}, 0)
	require.NoError(t, err)
	return input
}

func Test_AddQuestionCommand_shouldAddQuestion_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource).Return(true, nil)

	questionAdder := newMockquestionAdder(t)
	questionAdder.On("Add", mock.Anything, fixtureWorkbookID, "word_fill", "What is Go?", []string{"lang:en"}, 0).Return(fixtureQuestionID, nil)

	cmd := questionusecase.NewAddQuestionCommand(questionAdder, authChecker)
	input := newAddQuestionInput(t)

	// when
	output, err := cmd.AddQuestion(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureQuestionID, output.QuestionID)
	assert.Equal(t, "word_fill", output.QuestionType)
	assert.Equal(t, "What is Go?", output.Content)
}

func Test_AddQuestionCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource).Return(false, nil)

	cmd := questionusecase.NewAddQuestionCommand(nil, authChecker)
	input := newAddQuestionInput(t)

	// when
	_, err = cmd.AddQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_AddQuestionCommand_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authErr := errors.New("auth unavailable")
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource).Return(false, authErr)

	cmd := questionusecase.NewAddQuestionCommand(nil, authChecker)
	input := newAddQuestionInput(t)

	// when
	_, err = cmd.AddQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}
