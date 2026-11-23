package audio

import "slices"

// MaxBatchLimit is the largest number of items one audio batch may list or reclaim.
const MaxBatchLimit = 200

// MaxSynthesisInputBytes is the Cloud TTS input limit for the Neural2 voices
// cocotola-audio-generator uses; other voices (e.g. Gemini-TTS) have different limits.
const MaxSynthesisInputBytes = 5000

// FitsSynthesisInput reports whether text is within MaxSynthesisInputBytes.
func FitsSynthesisInput(text string) bool {
	return len(text) <= MaxSynthesisInputBytes
}

// SupportedLangs returns the short language codes that have a TTS voice.
func SupportedLangs() []string {
	return []string{"ja", "en"}
}

// IsSupportedLang reports whether audio can be generated for the language.
func IsSupportedLang(lang string) bool {
	return slices.Contains(SupportedLangs(), lang)
}
