package usersetting_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

var fixtureUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000573")

type outcome struct {
	value any
	err   error
}

func result[T any](v *T, err error) outcome {
	return outcome{value: v, err: err}
}

func Test_constructors_shouldReturnValueWithGivenArguments_whenArgumentsAreValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      outcome
		expected any
	}{
		{"UpdateLanguageInput", result(usersetting.NewUpdateLanguageInput(fixtureUserID, "ja")),
			&usersetting.UpdateLanguageInput{AppUserID: fixtureUserID, Language: "ja"}},
		{"UpdateDailyGoalInput", result(usersetting.NewUpdateDailyGoalInput(fixtureUserID, 25)),
			&usersetting.UpdateDailyGoalInput{AppUserID: fixtureUserID, DailyGoal: 25}},
		{"UpdateTimezoneInput", result(usersetting.NewUpdateTimezoneInput(fixtureUserID, "America/Los_Angeles")),
			&usersetting.UpdateTimezoneInput{AppUserID: fixtureUserID, Timezone: "America/Los_Angeles"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when: tt.got is built from valid fixtures

			// then
			require.NoError(t, tt.got.err)
			assert.Equal(t, tt.expected, tt.got.value)
		})
	}
}

func Test_constructors_shouldReturnErrInvalidArgument_whenAppUserIDIsZero(t *testing.T) {
	t.Parallel()

	var zero domain.AppUserID

	tests := []struct {
		name string
		got  outcome
	}{
		{"UpdateLanguageInput", result(usersetting.NewUpdateLanguageInput(zero, "ja"))},
		{"UpdateDailyGoalInput", result(usersetting.NewUpdateDailyGoalInput(zero, 25))},
		{"UpdateTimezoneInput", result(usersetting.NewUpdateTimezoneInput(zero, "America/Los_Angeles"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when: tt.got is built with a zero app user ID

			// then
			require.ErrorIs(t, tt.got.err, domain.ErrInvalidArgument)
		})
	}
}
