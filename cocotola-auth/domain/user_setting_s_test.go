package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_NewUserSetting_shouldReturnUserSetting_whenValid(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 15, "Asia/Tokyo")

	// then
	require.NoError(t, err)
	assert.Equal(t, appUserID, setting.AppUserID())
	assert.Equal(t, 5, setting.MaxWorkbooks())
	assert.Equal(t, "ja", setting.Language())
	assert.Equal(t, 15, setting.DailyGoal())
	assert.Equal(t, "Asia/Tokyo", setting.Timezone())
	assert.Equal(t, 0, setting.Version())
}

func Test_NewUserSetting_shouldReturnError_whenAppUserIDIsZero(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewUserSetting(domain.AppUserID{}, 5, "ja", 10, "Asia/Tokyo")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenMaxWorkbooksIsZero(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, 0, "ja", 10, "Asia/Tokyo")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenMaxWorkbooksIsNegative(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, -1, "ja", 10, "Asia/Tokyo")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenLanguageIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		language string
	}{
		{name: "empty", language: ""},
		{name: "uppercase", language: "JA"},
		{name: "threeLetters", language: "jpn"},
		{name: "withDigit", language: "j1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

			// when
			_, err := domain.NewUserSetting(appUserID, 5, tt.language, 10, "Asia/Tokyo")

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUserSetting_shouldReturnError_whenDailyGoalIsBelowMin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dailyGoal int
	}{
		{name: "zero", dailyGoal: 0},
		{name: "negative", dailyGoal: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

			// when
			_, err := domain.NewUserSetting(appUserID, 5, "ja", tt.dailyGoal, "Asia/Tokyo")

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUserSetting_shouldReturnError_whenDailyGoalExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, 5, "ja", 501, "Asia/Tokyo")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldSucceed_whenDailyGoalIsAtLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 500, "Asia/Tokyo")

	// then
	require.NoError(t, err)
	assert.Equal(t, 500, setting.DailyGoal())
}

func Test_NewUserSetting_shouldReturnError_whenTimezoneIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		timezone string
	}{
		{name: "empty", timezone: ""},
		{name: "garbage", timezone: "Not/AZone"},
		{name: "casing", timezone: "asia/tokyo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

			// when
			_, err := domain.NewUserSetting(appUserID, 5, "ja", 10, tt.timezone)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUserSetting_shouldSucceed_whenTimezoneIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		timezone string
	}{
		{name: "asiaTokyo", timezone: "Asia/Tokyo"},
		{name: "utc", timezone: "UTC"},
		{name: "americaLosAngeles", timezone: "America/Los_Angeles"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

			// when
			setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, tt.timezone)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.timezone, setting.Timezone())
		})
	}
}

func Test_NewDefaultUserSetting_shouldSetMaxWorkbooksTo3(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewDefaultUserSetting(appUserID)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, setting.MaxWorkbooks())
}

func Test_NewDefaultUserSetting_shouldSetLanguageToEn(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewDefaultUserSetting(appUserID)

	// then
	require.NoError(t, err)
	assert.Equal(t, "en", setting.Language())
}

func Test_NewDefaultUserSetting_shouldSetDailyGoalTo10(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewDefaultUserSetting(appUserID)

	// then
	require.NoError(t, err)
	assert.Equal(t, 10, setting.DailyGoal())
}

func Test_NewDefaultUserSetting_shouldSetTimezoneToAsiaTokyo(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewDefaultUserSetting(appUserID)

	// then
	require.NoError(t, err)
	assert.Equal(t, "Asia/Tokyo", setting.Timezone())
}

func Test_DefaultLanguage_shouldReturnEn(t *testing.T) {
	t.Parallel()

	// when
	lang := domain.DefaultLanguage()

	// then
	assert.Equal(t, "en", lang)
}

func Test_DefaultDailyGoal_shouldReturn10(t *testing.T) {
	t.Parallel()

	// when
	goal := domain.DefaultDailyGoal()

	// then
	assert.Equal(t, 10, goal)
}

func Test_DefaultTimezone_shouldReturnAsiaTokyo(t *testing.T) {
	t.Parallel()

	// when
	tz := domain.DefaultTimezone()

	// then
	assert.Equal(t, "Asia/Tokyo", tz)
}

func Test_ReconstructUserSetting_shouldReturnUserSetting_whenValid(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.ReconstructUserSetting(appUserID, 10, "en", 25, "UTC")

	// then
	require.NoError(t, err)
	assert.Equal(t, 10, setting.MaxWorkbooks())
	assert.Equal(t, "en", setting.Language())
	assert.Equal(t, 25, setting.DailyGoal())
	assert.Equal(t, "UTC", setting.Timezone())
	assert.Equal(t, 0, setting.Version())
}

func Test_ReconstructUserSetting_shouldReturnError_whenMaxWorkbooksIsZero(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.ReconstructUserSetting(appUserID, 0, "en", 10, "Asia/Tokyo")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenMaxWorkbooksExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, 101, "ja", 10, "Asia/Tokyo")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldSucceed_whenMaxWorkbooksIsAtLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewUserSetting(appUserID, 100, "ja", 10, "Asia/Tokyo")

	// then
	require.NoError(t, err)
	assert.Equal(t, 100, setting.MaxWorkbooks())
}

func Test_UserSetting_SetVersion_shouldSetVersion(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	setting.SetVersion(3)

	// then
	assert.Equal(t, 3, setting.Version())
}

func Test_UserSetting_ChangeLanguage_shouldUpdateLanguage(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeLanguage("en")

	// then
	require.NoError(t, err)
	assert.Equal(t, "en", setting.Language())
}

func Test_UserSetting_ChangeLanguage_shouldReturnError_whenLanguageIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeLanguage("INVALID")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_UserSetting_ChangeDailyGoal_shouldUpdateDailyGoal(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeDailyGoal(42)

	// then
	require.NoError(t, err)
	assert.Equal(t, 42, setting.DailyGoal())
}

func Test_UserSetting_ChangeDailyGoal_shouldReturnError_whenDailyGoalIsBelowMin(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeDailyGoal(0)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_UserSetting_ChangeDailyGoal_shouldReturnError_whenDailyGoalExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeDailyGoal(501)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_UserSetting_ChangeTimezone_shouldUpdateTimezone(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeTimezone("America/Los_Angeles")

	// then
	require.NoError(t, err)
	assert.Equal(t, "America/Los_Angeles", setting.Timezone())
}

func Test_UserSetting_ChangeTimezone_shouldReturnError_whenTimezoneIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, err := domain.NewUserSetting(appUserID, 5, "ja", 10, "Asia/Tokyo")
	require.NoError(t, err)

	// when
	err = setting.ChangeTimezone("Not/AZone")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
