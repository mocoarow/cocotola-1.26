package audio

import "slices"

// MaxBatchLimit is the largest number of items one audio batch may list or reclaim.
const MaxBatchLimit = 200

// SupportedLangs returns the short language codes that have a TTS voice.
func SupportedLangs() []string {
	return []string{"ja", "en"}
}

// IsSupportedLang reports whether audio can be generated for the language.
func IsSupportedLang(lang string) bool {
	return slices.Contains(SupportedLangs(), lang)
}
