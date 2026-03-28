package space_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
)

func Test_NewSpaceType_shouldReturnPublic_whenValueIsPublic(t *testing.T) {
	t.Parallel()

	// when
	st, err := space.NewType("public")

	// then
	require.NoError(t, err)
	assert.True(t, st.IsPublic())
	assert.False(t, st.IsPrivate())
	assert.Equal(t, "public", st.Value())
}

func Test_NewSpaceType_shouldReturnPrivate_whenValueIsPrivate(t *testing.T) {
	t.Parallel()

	// when
	st, err := space.NewType("private")

	// then
	require.NoError(t, err)
	assert.False(t, st.IsPublic())
	assert.True(t, st.IsPrivate())
	assert.Equal(t, "private", st.Value())
}

func Test_NewSpaceType_shouldReturnError_whenValueIsInvalid(t *testing.T) {
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
			_, err := space.NewType(tt.value)

			// then
			require.Error(t, err)
		})
	}
}

func Test_SpaceTypePublic_shouldReturnPublicType(t *testing.T) {
	t.Parallel()

	// when
	st := space.TypePublic()

	// then
	assert.True(t, st.IsPublic())
	assert.Equal(t, "public", st.Value())
}

func Test_SpaceTypePrivate_shouldReturnPrivateType(t *testing.T) {
	t.Parallel()

	// when
	st := space.TypePrivate()

	// then
	assert.True(t, st.IsPrivate())
	assert.Equal(t, "private", st.Value())
}
