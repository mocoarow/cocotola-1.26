package question_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func newReadyAudioQuestion(t *testing.T) *question.Question {
	t.Helper()
	a := validQuestionArgs()
	q, err := question.NewQuestion(contentjson.Parser{}, a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGenerationWithSourceRef(t, question.AudioGenerationStatusReady()))
	return q
}

func Test_Question_SyncWordFillAudio_shouldMarkPending_whenBothLangsHaveVoice(t *testing.T) {
	t.Parallel()

	// given
	q := newReadyAudioQuestion(t)
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんご", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{apple}}", Lang: "en"},
	}
	now := time.Now()

	// when
	q.SyncWordFillAudio(c, now)

	// then
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "pending", ag.Status().Value())
	assert.Equal(t, question.ComputeWordFillAudioInputHash(c), ag.InputHash())
}

func Test_Question_SyncWordFillAudio_shouldDiscardAudio_whenLangHasNoVoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sourceLang string
		targetLang string
	}{
		{name: "source without voice", sourceLang: "fr", targetLang: "en"},
		{name: "target without voice", sourceLang: "ja", targetLang: "fr"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: audio already generated for an earlier edit
			q := newReadyAudioQuestion(t)
			c := question.WordFillContent{
				Source: question.TextWithLang{Text: "pomme", Lang: tt.sourceLang},
				Target: question.TextWithLang{Text: "{{apple}}", Lang: tt.targetLang},
			}

			// when
			q.SyncWordFillAudio(c, time.Now())

			// then
			assert.Nil(t, q.AudioGeneration())
		})
	}
}
