package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func Test_NewQuestionType_shouldReturnDefault_whenValueIsDefault(t *testing.T) {
	t.Parallel()

	// when
	qt, err := question.NewType("default")

	// then
	require.NoError(t, err)
	assert.Equal(t, "default", qt.Value())
}

func Test_NewQuestionType_shouldReturnError_whenValueIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "unknown", value: "unknown"},
		{name: "uppercase", value: "Default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := question.NewType(tt.value)

			// then
			require.Error(t, err)
		})
	}
}

func Test_TypeDefault_shouldReturnDefaultType(t *testing.T) {
	t.Parallel()

	// when
	qt := question.TypeDefault()

	// then
	assert.Equal(t, "default", qt.Value())
}
