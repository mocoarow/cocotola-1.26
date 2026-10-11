package study_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func Test_GetDashboardInput_DaysValidatorTagMatchesConstants(t *testing.T) {
	t.Parallel()

	// given
	field, ok := reflect.TypeFor[studyservice.GetDashboardInput]().FieldByName("Days")
	require.True(t, ok, "GetDashboardInput.Days field must exist")
	tag := field.Tag.Get("validate")

	// when
	expected := fmt.Sprintf("gte=%d,lte=%d", studyservice.MinDashboardDays, studyservice.MaxDashboardDays)

	// then
	assert.Equal(t, expected, tag, "validate tag must mirror MinDashboardDays / MaxDashboardDays")
}
