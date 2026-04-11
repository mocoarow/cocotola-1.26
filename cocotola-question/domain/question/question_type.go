// Package question contains the question entity of the cocotola-question domain.
package question

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// Type represents the type of a question.
type Type struct {
	value string
}

const (
	questionTypeDefault = "default"
)

// TypeDefault returns the default question type.
func TypeDefault() Type { return Type{value: questionTypeDefault} }

// NewType creates a validated QuestionType from a string.
func NewType(value string) (Type, error) {
	switch value {
	case questionTypeDefault:
		return TypeDefault(), nil
	default:
		return Type{}, fmt.Errorf("invalid question type: must be 'default': %w", domain.ErrInvalidArgument)
	}
}

// Value returns the string representation.
func (t Type) Value() string { return t.value }
