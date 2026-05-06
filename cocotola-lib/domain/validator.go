package domain

import "github.com/go-playground/validator/v10"

var (
	v = validator.New() //nolint:gochecknoglobals
)

// ValidateStruct validates the given struct using the go-playground/validator tags.
func ValidateStruct(s any) error {
	return v.Struct(s) //nolint:wrapcheck
}
