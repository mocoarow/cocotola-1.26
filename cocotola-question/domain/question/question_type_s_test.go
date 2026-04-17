package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func Test_NewQuestionType_shouldReturnWordFill_whenValueIsWordFill(t *testing.T) {
	t.Parallel()

	// when
	qt, err := question.NewType("word_fill")

	// then
	require.NoError(t, err)
	assert.Equal(t, "word_fill", qt.Value())
}

func Test_NewQuestionType_shouldReturnMultipleChoice_whenValueIsMultipleChoice(t *testing.T) {
	t.Parallel()

	// when
	qt, err := question.NewType("multiple_choice")

	// then
	require.NoError(t, err)
	assert.Equal(t, "multiple_choice", qt.Value())
}

func Test_NewQuestionType_shouldReturnError_whenValueIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "unknown", value: "unknown"},
		{name: "default", value: "default"},
		{name: "uppercase", value: "Word_Fill"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := question.NewType(tt.value)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
