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
	questionTypeWordFill       = "word_fill"
	questionTypeMultipleChoice = "multiple_choice"
	questionTypeOrdering       = "ordering"
)

// TypeWordFill returns the word fill question type.
func TypeWordFill() Type { return Type{value: questionTypeWordFill} }

// TypeMultipleChoice returns the multiple choice question type.
func TypeMultipleChoice() Type { return Type{value: questionTypeMultipleChoice} }

// TypeOrdering returns the ordering question type.
func TypeOrdering() Type { return Type{value: questionTypeOrdering} }

// NewType creates a validated QuestionType from a string.
func NewType(value string) (Type, error) {
	switch value {
	case questionTypeWordFill:
		return TypeWordFill(), nil
	case questionTypeMultipleChoice:
		return TypeMultipleChoice(), nil
	case questionTypeOrdering:
		return TypeOrdering(), nil
	default:
		return Type{}, fmt.Errorf("invalid question type: must be 'word_fill', 'multiple_choice' or 'ordering': %w", domain.ErrInvalidArgument)
	}
}

// Value returns the string representation.
func (t Type) Value() string { return t.value }
