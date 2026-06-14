package study_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// Go struct tags cannot reference constants — the `validate:"gte=7,lte=730"`
// string on GetDashboardInput.Days must duplicate the MinDashboardDays /
// MaxDashboardDays values. This test fails fast when the two drift apart
// (e.g. someone bumps MaxDashboardDays without remembering to also touch
// the tag), so the bound stays the single source of truth at review time
// even though the compiler cannot enforce it.
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
