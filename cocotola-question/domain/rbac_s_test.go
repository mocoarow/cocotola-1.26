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
