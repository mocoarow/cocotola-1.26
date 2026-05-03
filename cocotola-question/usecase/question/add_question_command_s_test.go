package question_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

const (
	fixtureOperatorID      = "user-1"
	fixtureOrganizationID  = "org-1"
	fixtureWorkbookID      = "wb-1"
	fixtureQuestionID      = "q-1"
	fixtureWordFillContent = `{"source":{"text":"apple","lang":"en"},"target":{"text":"{{りんご}}","lang":"ja"}}`
)

func fixtureActiveQuestionList(t *testing.T, questionIDs ...string) *domain.ActiveQuestionList {
	t.Helper()
	list, err := domain.NewActiveQuestionList(fixtureWorkbookID, questionIDs)
	require.NoError(t, err)
	return list
}

func newAddQuestionInput(t *testing.T) *questionservice.AddQuestionInput {
	t.Helper()
	input, err := questionservice.NewAddQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, "word_fill", fixtureWordFillContent, []string{"lang:en"}, 0)
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

	questionSaver := newMockquestionSaver(t)
	questionSaver.On("Save", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		return q != nil &&
			q.WorkbookID() == fixtureWorkbookID &&
			q.QuestionType().Value() == "word_fill" &&
			q.Content() == fixtureWordFillContent &&
			q.OrderIndex() == 0 &&
			q.Version() == 0 &&
			len(q.Tags()) == 1 && q.Tags()[0] == "lang:en"
	})).Return(nil)

	activeListFinder := newMockactiveQuestionListFinder(t)
	activeListFinder.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t), nil)

	activeListSaver := newMockactiveQuestionListSaver(t)
	activeListSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	cmd := questionusecase.NewAddQuestionCommand(questionSaver, activeListFinder, activeListSaver, authChecker)
	input := newAddQuestionInput(t)

	// when
	output, err := cmd.AddQuestion(ctx, input)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, output.QuestionID)
	assert.Equal(t, "word_fill", output.QuestionType)
	assert.JSONEq(t, fixtureWordFillContent, output.Content)
	assert.Equal(t, []string{"lang:en"}, output.Tags)
	assert.Equal(t, 0, output.OrderIndex)
}

func Test_AddQuestionCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource).Return(false, nil)

	cmd := questionusecase.NewAddQuestionCommand(nil, nil, nil, authChecker)
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

	cmd := questionusecase.NewAddQuestionCommand(nil, nil, nil, authChecker)
	input := newAddQuestionInput(t)

	// when
	_, err = cmd.AddQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}
