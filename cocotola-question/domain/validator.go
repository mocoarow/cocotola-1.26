package domain

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	v = validator.New() //nolint:gochecknoglobals
)

// ValidateStruct validates the given struct using the go-playground/validator tags.
func ValidateStruct(s any) error {
	return v.Struct(s) //nolint:wrapcheck
}

// ValidateInput validates a client-supplied struct; a failure wraps ErrInvalidArgument.
func ValidateInput(s any) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	return nil
}
