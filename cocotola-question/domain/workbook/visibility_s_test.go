package workbook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

func Test_NewVisibility_shouldReturnPrivate_whenValueIsPrivate(t *testing.T) {
	t.Parallel()

	// when
	v, err := workbook.NewVisibility("private")

	// then
	require.NoError(t, err)
	assert.True(t, v.IsPrivate())
	assert.False(t, v.IsPublic())
	assert.Equal(t, "private", v.Value())
}

func Test_NewVisibility_shouldReturnPublic_whenValueIsPublic(t *testing.T) {
	t.Parallel()

	// when
	v, err := workbook.NewVisibility("public")

	// then
	require.NoError(t, err)
	assert.False(t, v.IsPrivate())
	assert.True(t, v.IsPublic())
	assert.Equal(t, "public", v.Value())
}

func Test_NewVisibility_shouldReturnError_whenValueIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "unknown", value: "unknown"},
		{name: "uppercase", value: "Public"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := workbook.NewVisibility(tt.value)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_VisibilityPrivate_shouldReturnPrivateType(t *testing.T) {
	t.Parallel()

	// when
	v := workbook.VisibilityPrivate()

	// then
	assert.True(t, v.IsPrivate())
	assert.Equal(t, "private", v.Value())
}

func Test_VisibilityPublic_shouldReturnPublicType(t *testing.T) {
	t.Parallel()

	// when
	v := workbook.VisibilityPublic()

	// then
	assert.True(t, v.IsPublic())
	assert.Equal(t, "public", v.Value())
}
