package i18n

import "regexp"

// LanguagePattern is the regular expression used across services to validate
// ISO 639-1 language codes. Codes are restricted to two lowercase ASCII
// letters (e.g. "ja", "en"). Higher-fidelity locale tags (e.g. "ja-JP") are
// intentionally not supported here.
var LanguagePattern = regexp.MustCompile(`^[a-z]{2}$`)

// IsValidISO6391 reports whether value is a lowercase two-letter ISO 639-1
// language code.
func IsValidISO6391(value string) bool {
	return LanguagePattern.MatchString(value)
}
