package question_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

const fixtureAudioInputHash = "a1b2c3d4e5f60718293a4b5c6d7e8f9001020304050607080910111213141516"

func newWordFillQuestionWithAudio(t *testing.T, id, workbookID string, status domainquestion.AudioGenerationStatus, inputHash string, updatedAt time.Time, failedAttempts int) domainquestion.Question {
	t.Helper()
	q, err := domainquestion.NewQuestion(
		id,
		workbookID,
		domainquestion.TypeWordFill(),
		fixtureWordFillContent,
		nil,
		0,
		updatedAt,
		updatedAt,
	)
	require.NoError(t, err)
	ag, err := domainquestion.NewAudioGeneration(status, inputHash, nil, updatedAt, failedAttempts, "")
	require.NoError(t, err)
	q.SetAudioGeneration(ag)
	return *q
}

func newMultipleChoiceQuestionWithAudio(t *testing.T, id, workbookID string, status domainquestion.AudioGenerationStatus, inputHash string, updatedAt time.Time) domainquestion.Question {
	t.Helper()
	content := `{"questionText":"Q?","choices":[{"id":"a","text":"A","isCorrect":true}],"displayCount":1}`
	q, err := domainquestion.NewQuestion(
		id,
		workbookID,
		domainquestion.TypeMultipleChoice(),
		content,
		nil,
		0,
		updatedAt,
		updatedAt,
	)
	require.NoError(t, err)
	ag, err := domainquestion.NewAudioGeneration(status, inputHash, nil, updatedAt, 0, "")
	require.NoError(t, err)
	q.SetAudioGeneration(ag)
	return *q
}

func Test_AudioBatchCommand_ListPendingAudio_shouldReturnItems_whenAllWordFill(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	now := time.Now()
	pending := newMockpendingAudioFinder(t)
	pending.On("FindPendingAudio", mock.Anything, 10).Return([]domainquestion.Question{
		newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHash, now, 0),
	}, nil)
	cmd := questionusecase.NewAudioBatchCommand(nil, pending, nil, nil, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewListPendingAudioInput(10)
	require.NoError(t, err)

	// when
	output, err := cmd.ListPendingAudio(ctx, input)

	// then
	require.NoError(t, err)
	require.Len(t, output.Items, 1)
	assert.Equal(t, "q-1", output.Items[0].QuestionID)
	assert.Equal(t, fixtureWorkbookID, output.Items[0].WorkbookID)
	assert.Equal(t, fixtureAudioInputHash, output.Items[0].InputHash)
	assert.Equal(t, "apple", output.Items[0].SourceText)
	assert.Equal(t, "りんご", output.Items[0].TargetText, "blank placeholder must be filled with inner word")
}

func Test_AudioBatchCommand_ListPendingAudio_shouldSkipItem_whenQuestionTypeIsNotWordFill(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	now := time.Now()
	pending := newMockpendingAudioFinder(t)
	pending.On("FindPendingAudio", mock.Anything, 10).Return([]domainquestion.Question{
		newWordFillQuestionWithAudio(t, "q-wf", fixtureWorkbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHash, now, 0),
		newMultipleChoiceQuestionWithAudio(t, "q-mc", fixtureWorkbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHash, now),
	}, nil)
	cmd := questionusecase.NewAudioBatchCommand(nil, pending, nil, nil, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewListPendingAudioInput(10)
	require.NoError(t, err)

	// when
	output, err := cmd.ListPendingAudio(ctx, input)

	// then
	require.NoError(t, err)
	require.Len(t, output.Items, 1)
	assert.Equal(t, "q-wf", output.Items[0].QuestionID)
}

func Test_AudioBatchCommand_ListPendingAudio_shouldSkipItem_whenAudioGenerationIsNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a question that somehow has no audio generation hydrated
	q, err := domainquestion.NewQuestion("q-1", fixtureWorkbookID, domainquestion.TypeWordFill(), fixtureWordFillContent, nil, 0, time.Now(), time.Now())
	require.NoError(t, err)
	pending := newMockpendingAudioFinder(t)
	pending.On("FindPendingAudio", mock.Anything, 10).Return([]domainquestion.Question{*q}, nil)
	cmd := questionusecase.NewAudioBatchCommand(nil, pending, nil, nil, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewListPendingAudioInput(10)
	require.NoError(t, err)

	// when
	output, err := cmd.ListPendingAudio(ctx, input)

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Items)
}

func Test_AudioBatchCommand_ListPendingAudio_shouldReturnError_whenFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finderErr := errors.New("firestore down")
	pending := newMockpendingAudioFinder(t)
	pending.On("FindPendingAudio", mock.Anything, 5).Return(nil, finderErr)
	cmd := questionusecase.NewAudioBatchCommand(nil, pending, nil, nil, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewListPendingAudioInput(5)
	require.NoError(t, err)

	// when
	_, err = cmd.ListPendingAudio(ctx, input)

	// then
	require.ErrorIs(t, err, finderErr)
}

func Test_AudioBatchCommand_ClaimAudio_shouldTransitionToGenerating_whenPending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	fixedNow := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		ag := q.AudioGeneration()
		return ag != nil && ag.Status().Value() == "generating" && ag.UpdatedAt().Equal(fixedNow)
	})).Return(nil)
	cfg := questionusecase.UsecaseConfig{ClockFunc: func() time.Time { return fixedNow }}
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, saver, cfg)
	input, err := questionservice.NewClaimAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash)
	require.NoError(t, err)

	// when
	err = cmd.ClaimAudio(ctx, input)

	// then
	require.NoError(t, err)
}

func Test_AudioBatchCommand_ClaimAudio_shouldReturnConcurrentModification_whenSaveLosesRace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewClaimAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash)
	require.NoError(t, err)

	// when
	err = cmd.ClaimAudio(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrAudioConcurrentModification)
}

func Test_AudioBatchCommand_ClaimAudio_shouldReturnError_whenStatusIsNotPending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, nil, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewClaimAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash)
	require.NoError(t, err)

	// when
	err = cmd.ClaimAudio(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrAudioNotPending)
}

func Test_AudioBatchCommand_CompleteAudio_shouldTransitionToReady_whenGenerating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	fixedNow := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		ag := q.AudioGeneration()
		return ag != nil && ag.Status().Value() == "ready" && ag.UpdatedAt().Equal(fixedNow)
	})).Return(nil)
	cfg := questionusecase.UsecaseConfig{ClockFunc: func() time.Time { return fixedNow }}
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, saver, cfg)
	input, err := questionservice.NewCompleteAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash, map[string]questionservice.CompleteAudioRefInput{
		domainquestion.AudioLangSource: {Path: "audio/questions/q-1/source.opus", DurationSec: 1.0, SizeBytes: 100},
		domainquestion.AudioLangTarget: {Path: "audio/questions/q-1/target.opus", DurationSec: 2.0, SizeBytes: 200},
	})
	require.NoError(t, err)

	// when
	err = cmd.CompleteAudio(ctx, input)

	// then
	require.NoError(t, err)
}

func Test_AudioBatchCommand_CompleteAudio_shouldReturnConcurrentModification_whenSaveLosesRace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewCompleteAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash, map[string]questionservice.CompleteAudioRefInput{
		domainquestion.AudioLangSource: {Path: "audio/questions/q-1/source.opus", DurationSec: 1.0, SizeBytes: 100},
	})
	require.NoError(t, err)

	// when
	err = cmd.CompleteAudio(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrAudioConcurrentModification)
}

func Test_AudioBatchCommand_FailAudio_shouldTransitionToFailed_whenGenerating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	fixedNow := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		ag := q.AudioGeneration()
		return ag != nil && ag.Status().Value() == "failed" && ag.LastError() == "tts: invalid voice" && ag.UpdatedAt().Equal(fixedNow)
	})).Return(nil)
	cfg := questionusecase.UsecaseConfig{ClockFunc: func() time.Time { return fixedNow }}
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, saver, cfg)
	input, err := questionservice.NewFailAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash, "tts: invalid voice")
	require.NoError(t, err)

	// when
	err = cmd.FailAudio(ctx, input)

	// then
	require.NoError(t, err)
}

func Test_AudioBatchCommand_FailAudio_shouldReturnConcurrentModification_whenSaveLosesRace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, time.Now(), 0)
	finder := newMockquestionFinder(t)
	finder.On("FindByID", mock.Anything, fixtureWorkbookID, "q-1").Return(&q, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)
	cmd := questionusecase.NewAudioBatchCommand(finder, nil, nil, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewFailAudioInput(fixtureWorkbookID, "q-1", fixtureAudioInputHash, "boom")
	require.NoError(t, err)

	// when
	err = cmd.FailAudio(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrAudioConcurrentModification)
}

func Test_AudioBatchCommand_ReclaimStaleAudio_shouldReclaimAllStaleItems(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: two stale generating items
	now := time.Now()
	stale := now.Add(-10 * time.Minute)
	q1 := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, stale, 0)
	q2 := newWordFillQuestionWithAudio(t, "q-2", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, stale, 0)
	finder := newMockstaleGeneratingAudioFinder(t)
	finder.On("FindStaleGenerating", mock.Anything, mock.Anything, 50).Return([]domainquestion.Question{q1, q2}, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.Anything).Return(nil).Twice()
	cmd := questionusecase.NewAudioBatchCommand(nil, nil, finder, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewReclaimStaleAudioInput(time.Minute, 50)
	require.NoError(t, err)

	// when
	output, err := cmd.ReclaimStaleAudio(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, output.Reclaimed)
}

func Test_AudioBatchCommand_ReclaimStaleAudio_shouldSkipItem_whenSaveLosesRace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: two stale items, the first save loses an optimistic-lock race
	now := time.Now()
	stale := now.Add(-10 * time.Minute)
	q1 := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, stale, 0)
	q2 := newWordFillQuestionWithAudio(t, "q-2", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, stale, 0)
	finder := newMockstaleGeneratingAudioFinder(t)
	finder.On("FindStaleGenerating", mock.Anything, mock.Anything, 50).Return([]domainquestion.Question{q1, q2}, nil)
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool { return q.ID() == "q-1" })).Return(libversioned.ErrConcurrentModification)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool { return q.ID() == "q-2" })).Return(nil)
	cmd := questionusecase.NewAudioBatchCommand(nil, nil, finder, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewReclaimStaleAudioInput(time.Minute, 50)
	require.NoError(t, err)

	// when
	output, err := cmd.ReclaimStaleAudio(ctx, input)

	// then: the loser is silently skipped, the winner counts
	require.NoError(t, err)
	assert.Equal(t, 1, output.Reclaimed)
}

func Test_AudioBatchCommand_ReclaimStaleAudio_shouldReturnError_whenSaveFailsForNonRaceReason(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a non-race save error
	now := time.Now()
	stale := now.Add(-10 * time.Minute)
	q1 := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHash, stale, 0)
	finder := newMockstaleGeneratingAudioFinder(t)
	finder.On("FindStaleGenerating", mock.Anything, mock.Anything, 50).Return([]domainquestion.Question{q1}, nil)
	saveErr := errors.New("firestore unavailable")
	saver := newMockquestionSaver(t)
	saver.On("Save", mock.Anything, mock.Anything).Return(saveErr)
	cmd := questionusecase.NewAudioBatchCommand(nil, nil, finder, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewReclaimStaleAudioInput(time.Minute, 50)
	require.NoError(t, err)

	// when
	output, err := cmd.ReclaimStaleAudio(ctx, input)

	// then
	require.ErrorIs(t, err, saveErr)
	assert.Equal(t, 0, output.Reclaimed)
}

func Test_AudioBatchCommand_ReclaimStaleAudio_shouldSkipItem_whenAudioGenerationStateIsNotStale(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a question whose status is no longer generating (race with another writer)
	now := time.Now()
	stale := now.Add(-10 * time.Minute)
	q := newWordFillQuestionWithAudio(t, "q-1", fixtureWorkbookID, domainquestion.AudioGenerationStatusReady(), fixtureAudioInputHash, stale, 0)
	finder := newMockstaleGeneratingAudioFinder(t)
	finder.On("FindStaleGenerating", mock.Anything, mock.Anything, 50).Return([]domainquestion.Question{q}, nil)
	saver := newMockquestionSaver(t)
	cmd := questionusecase.NewAudioBatchCommand(nil, nil, finder, saver, questionusecase.UsecaseConfig{})
	input, err := questionservice.NewReclaimStaleAudioInput(time.Minute, 50)
	require.NoError(t, err)

	// when
	output, err := cmd.ReclaimStaleAudio(ctx, input)

	// then: ReclaimStaleAudio domain method returns false → save is skipped, 0 reclaimed
	require.NoError(t, err)
	assert.Equal(t, 0, output.Reclaimed)
}
