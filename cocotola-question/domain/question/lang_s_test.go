package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func Test_NewLang_shouldReturnLang_whenCodeIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
	}{
		{name: "english", code: "en"},
		{name: "japanese", code: "ja"},
		{name: "italian", code: "it"},
		{name: "french", code: "fr"},
		{name: "german", code: "de"},
		{name: "spanish", code: "es"},
		{name: "chinese", code: "zh"},
		{name: "korean", code: "ko"},
		{name: "portuguese", code: "pt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			lang, err := question.NewLang(tt.code)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.code, lang.Value())
		})
	}
}

func Test_NewLang_shouldReturnError_whenCodeIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
	}{
		{name: "empty", code: ""},
		{name: "unknown", code: "xx"},
		{name: "uppercase", code: "EN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := question.NewLang(tt.code)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
