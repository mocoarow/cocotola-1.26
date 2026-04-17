package question

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// Lang represents an ISO 639-1 language code.
type Lang struct {
	value string
}

func isAllowedLanguage(code string) bool {
	switch code {
	case "en", "ja", "it", "fr", "de", "es", "zh", "ko", "pt":
		return true
	default:
		return false
	}
}

// NewLang creates a validated Lang from a string.
func NewLang(value string) (Lang, error) {
	if !isAllowedLanguage(value) {
		return Lang{}, fmt.Errorf("invalid language code %q: %w", value, domain.ErrInvalidArgument)
	}
	return Lang{value: value}, nil
}

// Value returns the string representation.
func (l Lang) Value() string { return l.value }
