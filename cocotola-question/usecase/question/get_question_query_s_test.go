package question_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

func newGetQuestionInput(t *testing.T) *questionservice.GetQuestionInput {
	t.Helper()
	in, err := questionservice.NewGetQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, fixtureQuestionID)
	require.NoError(t, err)
	return in
}

func newPublicWorkbook() *domainworkbook.Workbook {
	now := time.Now()
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, "space-1", fixtureOperatorID, fixtureOrganizationID, "title", "", domainworkbook.VisibilityPublic(), domainworkbook.LanguageJa(), 1, now, now)
}

func newPrivateWorkbook() *domainworkbook.Workbook {
	now := time.Now()
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, "space-1", fixtureOperatorID, fixtureOrganizationID, "title", "", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), 1, now, now)
}

func newFixtureQuestion() *domainquestion.Question {
	qType, _ := domainquestion.NewType("word_fill")
	now := time.Now()
	return domainquestion.ReconstructQuestion(fixtureQuestionID, fixtureWorkbookID, qType, fixtureWordFillContent, nil, 0, 1, now, now)
}

func newFixtureQuestionList() []domainquestion.Question {
	return []domainquestion.Question{*newFixtureQuestion()}
}

func Test_GetQuestionQuery_shouldReturnQuestion_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newPublicWorkbook(), nil)
	questionFinder := newMockquestionFinder(t)
	questionFinder.On("FindByID", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(newFixtureQuestion(), nil)
	authChecker := newMockauthorizationChecker(t) // no expectations: verifies IsAllowed is not called

	q := questionusecase.NewGetQuestionQuery(questionFinder, wbFinder, authChecker)

	// when
	output, err := q.GetQuestion(ctx, newGetQuestionInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureQuestionID, output.QuestionID)
}

func Test_GetQuestionQuery_shouldReturnForbidden_whenPrivateWorkbookAndNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newPrivateWorkbook(), nil)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource).Return(false, nil)

	q := questionusecase.NewGetQuestionQuery(nil, wbFinder, authChecker)

	// when
	_, err = q.GetQuestion(ctx, newGetQuestionInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_GetQuestionQuery_shouldReturnQuestion_whenPrivateWorkbookAndAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newPrivateWorkbook(), nil)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource).Return(true, nil)
	questionFinder := newMockquestionFinder(t)
	questionFinder.On("FindByID", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(newFixtureQuestion(), nil)

	q := questionusecase.NewGetQuestionQuery(questionFinder, wbFinder, authChecker)

	// when
	output, err := q.GetQuestion(ctx, newGetQuestionInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureQuestionID, output.QuestionID)
}

func Test_GetQuestionQuery_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	q := questionusecase.NewGetQuestionQuery(nil, wbFinder, nil)

	// when
	_, err := q.GetQuestion(ctx, newGetQuestionInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}
