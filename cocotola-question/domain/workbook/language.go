package workbook

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/i18n"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// Language represents the primary language of a workbook as an ISO 639-1 code.
type Language struct {
	value string
}

const (
	languageJa = "ja"
	languageEn = "en"
)

// LanguageJa returns the Japanese language.
func LanguageJa() Language { return Language{value: languageJa} }

// LanguageEn returns the English language.
func LanguageEn() Language { return Language{value: languageEn} }

// NewLanguage creates a validated Language from a string.
// The value must be a lowercase two-letter ISO 639-1 code (e.g. "ja", "en").
func NewLanguage(value string) (Language, error) {
	if !i18n.IsValidISO6391(value) {
		return Language{}, fmt.Errorf("invalid language %q: must be a lowercase ISO 639-1 code: %w", value, domain.ErrInvalidArgument)
	}
	return Language{value: value}, nil
}

// Value returns the string representation.
func (l Language) Value() string { return l.value }

// IsZero returns true if the language has no value set.
func (l Language) IsZero() bool { return l.value == "" }

// Equals reports whether two Language values are the same.
func (l Language) Equals(other Language) bool { return l.value == other.value }
