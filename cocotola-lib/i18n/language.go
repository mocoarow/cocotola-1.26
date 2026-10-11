package i18n

import "regexp"

// LanguagePattern validates ISO 639-1 codes; locale tags like "ja-JP" are
// intentionally rejected.
var LanguagePattern = regexp.MustCompile(`^[a-z]{2}$`)

// IsValidISO6391 reports whether value is an ISO 639-1 language code.
func IsValidISO6391(value string) bool {
	return LanguagePattern.MatchString(value)
}
