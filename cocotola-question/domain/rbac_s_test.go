package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func Test_NewAction_shouldReturnAction_whenValueIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "create_workbook", value: "create_workbook"},
		{name: "view_workbook", value: "view_workbook"},
		{name: "custom_action", value: "custom_action"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			action, err := domain.NewAction(tt.value)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.value, action.Value())
		})
	}
}

func Test_NewAction_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewAction("")

	// then
	assert.ErrorIs(t, err, domain.ErrEmptyActionValue)
}

func Test_NewResource_shouldReturnResource_whenValueIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "wildcard", value: "*"},
		{name: "specific", value: "workbook:123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			resource, err := domain.NewResource(tt.value)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.value, resource.Value())
		})
	}
}

func Test_NewResource_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewResource("")

	// then
	assert.ErrorIs(t, err, domain.ErrEmptyResourceValue)
}

func Test_ResourceSpace_shouldReturnResource_whenIDIsValid(t *testing.T) {
	t.Parallel()

	// when
	resource, err := domain.ResourceSpace("space-123")

	// then
	require.NoError(t, err)
	assert.Equal(t, "space:space-123", resource.Value())
}

func Test_ResourceSpace_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.ResourceSpace("")

	// then
	assert.ErrorIs(t, err, domain.ErrEmptyResourceValue)
}

func Test_ResourceWorkbook_shouldReturnResource_whenIDIsValid(t *testing.T) {
	t.Parallel()

	// when
	resource, err := domain.ResourceWorkbook("wb-123")

	// then
	require.NoError(t, err)
	assert.Equal(t, "workbook:wb-123", resource.Value())
}

func Test_ResourceWorkbook_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.ResourceWorkbook("")

	// then
	assert.ErrorIs(t, err, domain.ErrEmptyResourceValue)
}

func Test_NewEffect_shouldReturnEffect_whenValueIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "allow", value: "allow"},
		{name: "deny", value: "deny"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			effect, err := domain.NewEffect(tt.value)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.value, effect.Value())
		})
	}
}

func Test_NewEffect_shouldReturnError_whenValueIsInvalid(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewEffect("maybe")

	// then
	assert.ErrorIs(t, err, domain.ErrInvalidEffect)
}
