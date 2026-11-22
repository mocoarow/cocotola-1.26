package question_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
)

const fixtureNoVoiceWordFillContent = `{"source":{"text":"pomme","lang":"fr"},"target":{"text":"{{apple}}","lang":"en"}}`

func questionWithReadyAudio(t *testing.T) *domainquestion.Question {
	t.Helper()
	now := time.Now()
	q := domainquestion.ReconstructQuestion(fixtureQuestionID, fixtureWorkbookID, domainquestion.TypeWordFill(), fixtureWordFillContent, []string{"lang:en"}, 0, 1, now, now)
	ag, err := domainquestion.NewAudioGeneration(domainquestion.AudioGenerationStatusReady(), fixtureAudioInputHash, nil, now, 0, "")
	require.NoError(t, err)
	q.SetAudioGeneration(ag)
	return q
}

// newUpdateCommandOnReadyAudio returns a command whose Save mock only accepts
// a question satisfying matches.
func newUpdateCommandOnReadyAudio(t *testing.T, matches func(saved *domainquestion.Question) bool) *questionusecase.UpdateQuestionCommand {
	t.Helper()
	questionFinder := newMockquestionFinder(t)
	questionFinder.On("FindByID", mock.Anything, fixtureWorkbookID, fixtureQuestionID).Return(questionWithReadyAudio(t), nil)
	questionSaver := newMockquestionSaver(t)
	questionSaver.On("Save", mock.Anything, mock.MatchedBy(matches)).Return(nil)
	return questionusecase.NewUpdateQuestionCommand(questionFinder, questionSaver, existingWorkbookFinder(t), allowWorkbookAction(t, domain.ActionUpdateQuestion()), contentjson.Parser{})
}

func newUpdateInput(t *testing.T, content string) *questionservice.UpdateQuestionInput {
	t.Helper()
	input, err := questionservice.NewUpdateQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, fixtureQuestionID, content, []string{"lang:en"}, 0)
	require.NoError(t, err)
	return input
}

func Test_AddQuestionCommand_shouldQueueAudio_whenLangsHaveVoice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	questionCreator := newMockquestionCreator(t)
	questionCreator.On("Create", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		ag := q.AudioGeneration()
		return ag != nil && ag.Status().IsPending()
	})).Return(nil)
	cmd := questionusecase.NewAddQuestionCommand(questionCreator, existingWorkbookFinder(t), allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})

	// when
	_, err := cmd.AddQuestion(ctx, newAddQuestionInput(t))

	// then: Create is only matched when audio is queued
	require.NoError(t, err)
}

func Test_AddQuestionCommand_shouldNotQueueAudio_whenLangHasNoVoice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	questionCreator := newMockquestionCreator(t)
	questionCreator.On("Create", mock.Anything, mock.MatchedBy(func(q *domainquestion.Question) bool {
		return q.AudioGeneration() == nil
	})).Return(nil)
	cmd := questionusecase.NewAddQuestionCommand(questionCreator, existingWorkbookFinder(t), allowWorkbookAction(t, domain.ActionCreateQuestion()), contentjson.Parser{})
	input, err := questionservice.NewAddQuestionInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, "word_fill", fixtureNoVoiceWordFillContent, []string{"lang:en"}, 0)
	require.NoError(t, err)

	// when
	_, err = cmd.AddQuestion(ctx, input)

	// then: Create is only matched when no audio is queued
	require.NoError(t, err)
}

func Test_UpdateQuestionCommand_shouldQueueAudio_whenLangsHaveVoice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// given
	cmd := newUpdateCommandOnReadyAudio(t, func(saved *domainquestion.Question) bool {
		ag := saved.AudioGeneration()
		return ag != nil && ag.Status().IsPending()
	})

	// when
	_, err := cmd.UpdateQuestion(ctx, newUpdateInput(t, fixtureUpdatedWordFillContent))

	// then: Save is only matched when audio is pending again
	require.NoError(t, err)
}

func Test_UpdateQuestionCommand_shouldDiscardAudio_whenLangHasNoVoice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	cmd := newUpdateCommandOnReadyAudio(t, func(saved *domainquestion.Question) bool {
		return saved.AudioGeneration() == nil
	})

	// when
	_, err := cmd.UpdateQuestion(ctx, newUpdateInput(t, fixtureNoVoiceWordFillContent))

	// then: Save is only matched when the ready audio is gone
	require.NoError(t, err)
}
