package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/i18n"
)

func Test_IsValidISO6391_shouldReturnTrue_whenValueIsLowercaseTwoLetters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "ja", value: "ja"},
		{name: "en", value: "en"},
		{name: "fr", value: "fr"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			ok := i18n.IsValidISO6391(tt.value)

			// then
			assert.True(t, ok)
		})
	}
}

func Test_IsValidISO6391_shouldReturnFalse_whenValueIsInvalid(t *testing.T) {
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
		{name: "leadingSpace", value: " ja"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			ok := i18n.IsValidISO6391(tt.value)

			// then
			assert.False(t, ok)
		})
	}
}
