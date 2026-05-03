package question_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

const fixtureUpdatedWordFillContent = `{"source":{"text":"hello","lang":"en"},"target":{"text":"{{ハロー}}","lang":"ja"}}`

func newUpdateQuestionInput(t *testing.T) *questionservice.UpdateQuestionInput {
	t.Helper()
	input, err := questionservice.NewUpdateQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, fixtureQuestionID, fixtureUpdatedWordFillContent, []string{"lang:ja"}, 1)
	require.NoError(t, err)
	return input
}

func Test_UpdateQuestionCommand_shouldUpdateQuestion_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateQuestion(), wbResource).Return(true, nil)

	now := time.Now()
	qType, _ := domainquestion.NewType("word_fill")
	q := domainquestion.ReconstructQuestion(fixtureQuestionID, fixtureWorkbookID, qType, "Original content", []string{"lang:en"}, 0, 1, now, now)

	questionFinder := newMockquestionFinder(t)
	questionFinder.On("FindByID", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(q, nil)

	questionSaver := newMockquestionSaver(t)
	questionSaver.On("Save", mock.Anything, mock.MatchedBy(func(saved *domainquestion.Question) bool {
		return saved != nil &&
			saved.ID() == fixtureQuestionID &&
			saved.WorkbookID() == fixtureWorkbookID &&
			saved.Content() == fixtureUpdatedWordFillContent &&
			saved.OrderIndex() == 1 &&
			len(saved.Tags()) == 1 && saved.Tags()[0] == "lang:ja"
	})).Return(nil)

	cmd := questionusecase.NewUpdateQuestionCommand(questionFinder, questionSaver, authChecker)
	input := newUpdateQuestionInput(t)

	// when
	output, err := cmd.UpdateQuestion(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureQuestionID, output.QuestionID)
	assert.JSONEq(t, fixtureUpdatedWordFillContent, output.Content)
	assert.Equal(t, []string{"lang:ja"}, output.Tags)
	assert.Equal(t, 1, output.OrderIndex)
}

func Test_UpdateQuestionCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateQuestion(), wbResource).Return(false, nil)

	cmd := questionusecase.NewUpdateQuestionCommand(nil, nil, authChecker)
	input := newUpdateQuestionInput(t)

	// when
	_, err = cmd.UpdateQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_UpdateQuestionCommand_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authErr := errors.New("auth unavailable")
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateQuestion(), wbResource).Return(false, authErr)

	cmd := questionusecase.NewUpdateQuestionCommand(nil, nil, authChecker)
	input := newUpdateQuestionInput(t)

	// when
	_, err = cmd.UpdateQuestion(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}
