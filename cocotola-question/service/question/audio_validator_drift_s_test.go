package question_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/audio"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

func Test_ListPendingAndReclaimStaleAudioInput_LimitValidatorTagMatchesMaxBatchLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input reflect.Type
	}{
		{name: "ListPendingAudioInput", input: reflect.TypeFor[questionservice.ListPendingAudioInput]()},
		{name: "ReclaimStaleAudioInput", input: reflect.TypeFor[questionservice.ReclaimStaleAudioInput]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			field, ok := tt.input.FieldByName("Limit")
			require.True(t, ok, "Limit field must exist")
			tag := field.Tag.Get("validate")

			// when
			expected := fmt.Sprintf("required,gte=1,lte=%d", audio.MaxBatchLimit)

			// then
			assert.Equal(t, expected, tag, "validate tag must mirror audio.MaxBatchLimit")
		})
	}
}
