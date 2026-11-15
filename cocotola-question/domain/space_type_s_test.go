package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func Test_NewSpaceType_shouldReturnSpaceType_whenValueIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  domain.SpaceType
	}{
		{name: "public", value: "public", want: domain.SpaceTypePublic()},
		{name: "private", value: "private", want: domain.SpaceTypePrivate()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			got, err := domain.NewSpaceType(tt.value)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewSpaceType_shouldReturnErrInvalidSpaceType_whenValueIsUnknown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "unknown", value: "personal"},
		{name: "uppercase", value: "Public"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := domain.NewSpaceType(tt.value)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidSpaceType)
		})
	}
}
