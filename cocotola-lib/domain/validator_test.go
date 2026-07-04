package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
)

type validStruct struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=0,lte=150"`
}

type requiredFieldStruct struct {
	Name string `validate:"required"`
}

type rangeFieldStruct struct {
	Score int `validate:"gte=0,lte=100"`
}

func Test_ValidateStruct_shouldReturnNil_whenStructIsValid(t *testing.T) {
	t.Parallel()

	// given
	s := validStruct{Name: "Alice", Age: 30}

	// when
	err := domain.ValidateStruct(s)

	// then
	require.NoError(t, err)
}

func Test_ValidateStruct_shouldReturnError_whenRequiredFieldIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	s := requiredFieldStruct{Name: ""}

	// when
	err := domain.ValidateStruct(s)

	// then
	assert.Error(t, err)
}

func Test_ValidateStruct_shouldReturnError_whenFieldFailsValidationTag(t *testing.T) {
	t.Parallel()

	// given
	s := rangeFieldStruct{Score: 200}

	// when
	err := domain.ValidateStruct(s)

	// then
	assert.Error(t, err)
}
