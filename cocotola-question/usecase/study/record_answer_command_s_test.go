package study_test

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
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
)

const (
	fixtureOperatorID     = "user-1"
	fixtureOrganizationID = "org-1"
	fixtureWorkbookID     = "wb-1"
	fixtureQuestionID     = "q-1"
)

var fixtureClock = time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

var noShuffle = func(_ int, _ func(i, j int)) {}

var testConfig = studyusecase.UsecaseConfig{
	ClockFunc:   func() time.Time { return fixtureClock },
	ShuffleFunc: noShuffle,
}

func fixtureWorkbook() *domainworkbook.Workbook {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, "space-1", fixtureOperatorID, fixtureOrganizationID, "Test WB", "desc", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), now, now)
}

func fixturePublicWorkbook() *domainworkbook.Workbook {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, "public-space-1", "system-owner", fixtureOrganizationID, "Public WB", "desc", domainworkbook.VisibilityPublic(), domainworkbook.LanguageJa(), now, now)
}

func fixtureQuestions() []domainquestion.Question {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	return []domainquestion.Question{
		*domainquestion.ReconstructQuestion(fixtureQuestionID, domainquestion.TypeWordFill(), `{"source":"hello","target":"こんにちは","sourceLang":"en","targetLang":"ja","blanks":["hello"]}`, nil, 0, now, now),
	}
}

func fixtureActiveQuestionList(t *testing.T, questionIDs ...string) *domain.ActiveQuestionList {
	t.Helper()
	list, err := domain.NewActiveQuestionList(fixtureWorkbookID, questionIDs)
	require.NoError(t, err)
	return list
}

func newRecordAnswerInput(t *testing.T, correct bool) *studyservice.RecordAnswerInput {
	t.Helper()
	input, err := studyservice.NewRecordAnswerInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, fixtureQuestionID, correct)
	require.NoError(t, err)
	return input
}

func Test_RecordAnswerCommand_shouldRecordCorrectAnswer_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	finder := newMockstudyRecordFinder(t)
	finder.On("FindByID", mock.Anything, fixtureOperatorID, fixtureWorkbookID, fixtureQuestionID).Return(nil, domain.ErrStudyRecordNotFound)

	saver := newMockstudyRecordSaver(t)
	saver.On("Save", mock.Anything, fixtureOperatorID, mock.AnythingOfType("*study.Record")).Return(nil)

	cmd := studyusecase.NewRecordAnswerCommand(finder, saver, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newRecordAnswerInput(t, true)

	// when
	output, err := cmd.RecordAnswer(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, output.ConsecutiveCorrect)
	assert.Equal(t, 1, output.TotalCorrect)
	assert.Equal(t, 0, output.TotalIncorrect)
}

func Test_RecordAnswerCommand_shouldRecordIncorrectAnswer_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	existingRecord := domainstudy.ReconstructRecord(fixtureWorkbookID, fixtureQuestionID, 3, now, now, 3, 0)

	finder := newMockstudyRecordFinder(t)
	finder.On("FindByID", mock.Anything, fixtureOperatorID, fixtureWorkbookID, fixtureQuestionID).Return(existingRecord, nil)

	saver := newMockstudyRecordSaver(t)
	saver.On("Save", mock.Anything, fixtureOperatorID, mock.AnythingOfType("*study.Record")).Return(nil)

	cmd := studyusecase.NewRecordAnswerCommand(finder, saver, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newRecordAnswerInput(t, false)

	// when
	output, err := cmd.RecordAnswer(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, output.ConsecutiveCorrect)
	assert.Equal(t, 3, output.TotalCorrect)
	assert.Equal(t, 1, output.TotalIncorrect)
}

func Test_RecordAnswerCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, nil)

	cmd := studyusecase.NewRecordAnswerCommand(nil, nil, nil, workbookRepo, authChecker, testConfig)
	input := newRecordAnswerInput(t, true)

	// when
	_, err = cmd.RecordAnswer(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_RecordAnswerCommand_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	cmd := studyusecase.NewRecordAnswerCommand(nil, nil, nil, workbookRepo, nil, testConfig)
	input := newRecordAnswerInput(t, true)

	// when
	_, err := cmd.RecordAnswer(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}

func Test_RecordAnswerCommand_shouldReturnError_whenQuestionNotInWorkbook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(true, nil)

	// Active list does NOT contain fixtureQuestionID
	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, "other-q"), nil)

	cmd := studyusecase.NewRecordAnswerCommand(nil, nil, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newRecordAnswerInput(t, true)

	// when
	_, err = cmd.RecordAnswer(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrQuestionNotFound)
}

func Test_RecordAnswerCommand_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixtureWorkbook(), nil)

	authErr := errors.New("auth unavailable")
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionStudyWorkbook(), wbResource).Return(false, authErr)

	cmd := studyusecase.NewRecordAnswerCommand(nil, nil, nil, workbookRepo, authChecker, testConfig)
	input := newRecordAnswerInput(t, true)

	// when
	_, err = cmd.RecordAnswer(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}

func Test_RecordAnswerCommand_shouldRecordAnswer_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	workbookRepo := newMockworkbookFinder(t)
	workbookRepo.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixturePublicWorkbook(), nil)

	// authChecker should NOT be called for public workbooks.
	authChecker := newMockauthorizationChecker(t)

	activeListRepo := newMockactiveQuestionListFinder(t)
	activeListRepo.On("FindByWorkbookID", mock.Anything, fixtureWorkbookID).Return(fixtureActiveQuestionList(t, fixtureQuestionID), nil)

	finder := newMockstudyRecordFinder(t)
	finder.On("FindByID", mock.Anything, fixtureOperatorID, fixtureWorkbookID, fixtureQuestionID).Return(nil, domain.ErrStudyRecordNotFound)

	saver := newMockstudyRecordSaver(t)
	saver.On("Save", mock.Anything, fixtureOperatorID, mock.AnythingOfType("*study.Record")).Return(nil)

	cmd := studyusecase.NewRecordAnswerCommand(finder, saver, activeListRepo, workbookRepo, authChecker, testConfig)
	input := newRecordAnswerInput(t, true)

	// when
	output, err := cmd.RecordAnswer(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, output.ConsecutiveCorrect)
	assert.Equal(t, 1, output.TotalCorrect)
	assert.Equal(t, 0, output.TotalIncorrect)
	authChecker.AssertNotCalled(t, "IsAllowed")
}
