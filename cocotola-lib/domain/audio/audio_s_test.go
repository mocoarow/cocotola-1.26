package audio_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/audio"
)

func Test_SupportedLangs_shouldReturnJaAndEn(t *testing.T) {
	t.Parallel()

	// given / when
	got := audio.SupportedLangs()

	// then
	assert.Equal(t, []string{"ja", "en"}, got)
}

func Test_IsSupportedLang_shouldReturnTrue_whenLangIsInSupportedLangs(t *testing.T) {
	t.Parallel()

	for _, lang := range audio.SupportedLangs() {
		t.Run(lang, func(t *testing.T) {
			t.Parallel()

			// given: a language listed in SupportedLangs

			// when
			got := audio.IsSupportedLang(lang)

			// then
			assert.True(t, got)
		})
	}
}

func Test_IsSupportedLang_shouldReturnFalse_whenLangIsNotSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		lang string
	}{
		{name: "language without a voice", lang: "fr"},
		{name: "empty code", lang: ""},
		{name: "uppercase code", lang: "JA"},
		{name: "full locale", lang: "ja-JP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: a code that is not a supported short language code

			// when
			got := audio.IsSupportedLang(tt.lang)

			// then
			assert.False(t, got)
		})
	}
}
