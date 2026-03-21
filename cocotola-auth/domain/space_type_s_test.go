package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_NewSpaceType_shouldReturnPublic_whenValueIsPublic(t *testing.T) {
	t.Parallel()

	// when
	st, err := domain.NewSpaceType("public")

	// then
	require.NoError(t, err)
	assert.True(t, st.IsPublic())
	assert.False(t, st.IsPrivate())
	assert.Equal(t, "public", st.Value())
}

func Test_NewSpaceType_shouldReturnPrivate_whenValueIsPrivate(t *testing.T) {
	t.Parallel()

	// when
	st, err := domain.NewSpaceType("private")

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
			_, err := domain.NewSpaceType(tt.value)

			// then
			require.Error(t, err)
		})
	}
}

func Test_SpaceTypePublic_shouldReturnPublicType(t *testing.T) {
	t.Parallel()

	// when
	st := domain.SpaceTypePublic()

	// then
	assert.True(t, st.IsPublic())
	assert.Equal(t, "public", st.Value())
}

func Test_SpaceTypePrivate_shouldReturnPrivateType(t *testing.T) {
	t.Parallel()

	// when
	st := domain.SpaceTypePrivate()

	// then
	assert.True(t, st.IsPrivate())
	assert.Equal(t, "private", st.Value())
}
