package domain_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

type validatorTarget struct {
	Name string `validate:"required"`
}

func Test_ValidateInput_shouldReturnErrInvalidArgument_whenInputIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	input := &validatorTarget{}

	// when
	err := domain.ValidateInput(input)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateInput_shouldKeepValidationErrors_whenInputIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	input := &validatorTarget{}

	// when
	err := domain.ValidateInput(input)

	// then
	var verrs validator.ValidationErrors
	require.ErrorAs(t, err, &verrs)
}

func Test_ValidateInput_shouldReturnNil_whenInputIsValid(t *testing.T) {
	t.Parallel()

	// given
	input := &validatorTarget{Name: "name"}

	// when
	err := domain.ValidateInput(input)

	// then
	require.NoError(t, err)
}

func Test_ValidateStruct_shouldNotReturnErrInvalidArgument_whenStructIsInvalid(t *testing.T) {
	t.Parallel()

	// given: a server-built value, whose failure is not the client's fault
	output := &validatorTarget{}

	// when
	err := domain.ValidateStruct(output)

	// then
	var verrs validator.ValidationErrors
	require.ErrorAs(t, err, &verrs)
	assert.NotErrorIs(t, err, domain.ErrInvalidArgument)
}
