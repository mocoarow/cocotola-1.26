// Package question contains the question entity of the cocotola-question domain.
package question

import "errors"

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
		return Type{}, errors.New("invalid question type: must be 'default'")
	}
}

// Value returns the string representation.
func (t Type) Value() string { return t.value }
