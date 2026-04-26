package workbook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

func Test_NewLanguage_shouldReturnLanguage_whenValueIsJa(t *testing.T) {
	t.Parallel()

	// when
	l, err := workbook.NewLanguage("ja")

	// then
	require.NoError(t, err)
	assert.Equal(t, "ja", l.Value())
	assert.False(t, l.IsZero())
}

func Test_NewLanguage_shouldReturnLanguage_whenValueIsEn(t *testing.T) {
	t.Parallel()

	// when
	l, err := workbook.NewLanguage("en")

	// then
	require.NoError(t, err)
	assert.Equal(t, "en", l.Value())
}

func Test_NewLanguage_shouldReturnError_whenValueIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "uppercase", value: "JA"},
		{name: "mixedCase", value: "Ja"},
		{name: "threeLetters", value: "jpn"},
		{name: "oneLetter", value: "j"},
		{name: "withDigit", value: "j1"},
		{name: "withHyphen", value: "ja-JP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := workbook.NewLanguage(tt.value)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_LanguageJa_shouldReturnJa(t *testing.T) {
	t.Parallel()

	// when
	l := workbook.LanguageJa()

	// then
	assert.Equal(t, "ja", l.Value())
	assert.False(t, l.IsZero())
}

func Test_LanguageEn_shouldReturnEn(t *testing.T) {
	t.Parallel()

	// when
	l := workbook.LanguageEn()

	// then
	assert.Equal(t, "en", l.Value())
}

func Test_Language_IsZero_shouldReturnTrue_whenLanguageIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	var l workbook.Language

	// when
	zero := l.IsZero()

	// then
	assert.True(t, zero)
	assert.Empty(t, l.Value())
}

func Test_Language_Equals_shouldReturnTrue_whenLanguagesMatch(t *testing.T) {
	t.Parallel()

	// given
	a := workbook.LanguageJa()
	b, _ := workbook.NewLanguage("ja")

	// when
	equal := a.Equals(b)

	// then
	assert.True(t, equal)
}

func Test_Language_Equals_shouldReturnFalse_whenLanguagesDiffer(t *testing.T) {
	t.Parallel()

	// given
	a := workbook.LanguageJa()
	b := workbook.LanguageEn()

	// when
	equal := a.Equals(b)

	// then
	assert.False(t, equal)
}
