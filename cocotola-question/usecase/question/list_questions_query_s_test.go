package question_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

func newListQuestionsInput(t *testing.T) *questionservice.ListQuestionsInput {
	t.Helper()
	in, err := questionservice.NewListQuestionsInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID)
	require.NoError(t, err)
	return in
}

func Test_ListQuestionsQuery_shouldReturnQuestions_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newPublicWorkbook(), nil)
	questionFinder := newMockquestionFinder(t)
	questionFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(newFixtureQuestionList(), nil)
	authChecker := newMockauthorizationChecker(t) // no expectations: verifies IsAllowed is not called

	q := questionusecase.NewListQuestionsQuery(questionFinder, wbFinder, authChecker)

	// when
	output, err := q.ListQuestions(ctx, newListQuestionsInput(t))

	// then
	require.NoError(t, err)
	assert.Len(t, output.Questions, 1)
}

func Test_ListQuestionsQuery_shouldReturnForbidden_whenPrivateWorkbookAndNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newPrivateWorkbook(), nil)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource).Return(false, nil)

	q := questionusecase.NewListQuestionsQuery(nil, wbFinder, authChecker)

	// when
	_, err = q.ListQuestions(ctx, newListQuestionsInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ListQuestionsQuery_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	q := questionusecase.NewListQuestionsQuery(nil, wbFinder, nil)

	// when
	_, err := q.ListQuestions(ctx, newListQuestionsInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}
