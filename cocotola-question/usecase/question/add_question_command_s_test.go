package question_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
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

	questionCreator := newMockquestionCreator(t)
	questionCreator.On("Create", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		return q != nil &&
			q.WorkbookID() == fixtureWorkbookID &&
			q.QuestionType().Value() == "word_fill" &&
			q.Content() == fixtureWordFillContent &&
			q.OrderIndex() == 0 &&
			q.Version() == 0 &&
			len(q.Tags()) == 1 && q.Tags()[0] == "lang:en"
	})).Return(nil)

	cmd := questionusecase.NewAddQuestionCommand(questionCreator, authChecker, contentjson.Parser{})
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

	cmd := questionusecase.NewAddQuestionCommand(nil, authChecker, contentjson.Parser{})
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

	cmd := questionusecase.NewAddQuestionCommand(nil, authChecker, contentjson.Parser{})
	input := newAddQuestionInput(t)

	// when
	_, err = cmd.AddQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}

func Test_AddQuestionCommand_shouldReturnError_whenInjectedParserFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource).Return(true, nil)

	parseErr := errors.New("parse failed")
	parser := newMockcontentParser(t)
	parser.On("ParseWordFill", fixtureWordFillContent).Return(nil, parseErr)

	cmd := questionusecase.NewAddQuestionCommand(newMockquestionCreator(t), authChecker, parser)
	input := newAddQuestionInput(t)

	// when
	_, err = cmd.AddQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, parseErr)
}

func Test_AddQuestionCommand_shouldReturnError_whenCreateFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createErr := errors.New("firestore unavailable")
	questionCreator := newMockquestionCreator(t)
	questionCreator.On("Create", mock.Anything, mock.Anything).Return(createErr)

	cmd := questionusecase.NewAddQuestionCommand(questionCreator, allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, err := cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then
	require.ErrorIs(t, err, createErr)
}
