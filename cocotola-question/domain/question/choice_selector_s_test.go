package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func Test_DefaultChoiceSelector_shouldIncludeAllCorrectChoices(t *testing.T) {
	t.Parallel()

	// given
	choices := []question.ChoiceJSON{
		{ID: "c1", Text: "correct1", IsCorrect: true},
		{ID: "c2", Text: "correct2", IsCorrect: true},
		{ID: "c3", Text: "wrong1", IsCorrect: false},
		{ID: "c4", Text: "wrong2", IsCorrect: false},
		{ID: "c5", Text: "wrong3", IsCorrect: false},
	}

	// when
	result := question.DefaultChoiceSelector(choices, 3, false)

	// then
	require.Len(t, result, 3)

	correctIDs := make(map[string]bool)
	for _, c := range result {
		if c.IsCorrect {
			correctIDs[c.ID] = true
		}
	}
	assert.True(t, correctIDs["c1"])
	assert.True(t, correctIDs["c2"])
}

func Test_DefaultChoiceSelector_shouldShuffleWhenEnabled(t *testing.T) {
	t.Parallel()

	// given
	choices := []question.ChoiceJSON{
		{ID: "c1", Text: "correct", IsCorrect: true},
		{ID: "c2", Text: "wrong1", IsCorrect: false},
		{ID: "c3", Text: "wrong2", IsCorrect: false},
		{ID: "c4", Text: "wrong3", IsCorrect: false},
	}

	// when
	shuffled := false
	for range 10 {
		result := question.DefaultChoiceSelector(choices, 3, true)
		if result[0].ID != "c1" {
			shuffled = true
			break
		}
	}

	// then
	assert.True(t, shuffled, "expected at least one run to produce shuffled result")
}

func Test_DefaultChoiceSelector_shouldReturnExactDisplayCount(t *testing.T) {
	t.Parallel()

	// given
	choices := []question.ChoiceJSON{
		{ID: "c1", Text: "correct", IsCorrect: true},
		{ID: "c2", Text: "wrong1", IsCorrect: false},
		{ID: "c3", Text: "wrong2", IsCorrect: false},
		{ID: "c4", Text: "wrong3", IsCorrect: false},
		{ID: "c5", Text: "wrong4", IsCorrect: false},
	}

	// when
	result := question.DefaultChoiceSelector(choices, 3, false)

	// then
	assert.Len(t, result, 3)
}

func Test_DefaultChoiceSelector_shouldHandleAllCorrect(t *testing.T) {
	t.Parallel()

	// given
	choices := []question.ChoiceJSON{
		{ID: "c1", Text: "correct1", IsCorrect: true},
		{ID: "c2", Text: "correct2", IsCorrect: true},
	}

	// when
	result := question.DefaultChoiceSelector(choices, 2, false)

	// then
	assert.Len(t, result, 2)
	for _, c := range result {
		assert.True(t, c.IsCorrect)
	}
}
