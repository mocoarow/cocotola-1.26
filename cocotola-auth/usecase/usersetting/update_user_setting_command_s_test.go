package usersetting_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/usersetting"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

const (
	fixtureMaxWorkbooks = 5
	fixtureLanguage     = "en"
	fixtureDailyGoal    = 25
	fixtureTimezone     = "America/Los_Angeles"
	fixtureVersion      = 7
)

var fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000573")

type updateCall func(ctx context.Context, cmd *usersetting.UpdateUserSettingCommand) error

func updateLanguage(language string) updateCall {
	return func(ctx context.Context, cmd *usersetting.UpdateUserSettingCommand) error {
		return cmd.UpdateLanguage(ctx, &usersettingservice.UpdateLanguageInput{AppUserID: fixtureAppUserID, Language: language})
	}
}

func updateDailyGoal(dailyGoal int) updateCall {
	return func(ctx context.Context, cmd *usersetting.UpdateUserSettingCommand) error {
		return cmd.UpdateDailyGoal(ctx, &usersettingservice.UpdateDailyGoalInput{AppUserID: fixtureAppUserID, DailyGoal: dailyGoal})
	}
}

func updateTimezone(timezone string) updateCall {
	return func(ctx context.Context, cmd *usersetting.UpdateUserSettingCommand) error {
		return cmd.UpdateTimezone(ctx, &usersettingservice.UpdateTimezoneInput{AppUserID: fixtureAppUserID, Timezone: timezone})
	}
}

func newExistingSetting(t *testing.T) *domain.UserSetting {
	t.Helper()
	setting, err := domain.NewUserSetting(fixtureAppUserID, fixtureMaxWorkbooks, fixtureLanguage, fixtureDailyGoal, fixtureTimezone)
	require.NoError(t, err)
	setting.SetVersion(fixtureVersion)
	return setting
}

type settingValues struct {
	maxWorkbooks int
	language     string
	dailyGoal    int
	timezone     string
	version      int
}

func valuesOf(s *domain.UserSetting) settingValues {
	return settingValues{
		maxWorkbooks: s.MaxWorkbooks(),
		language:     s.Language(),
		dailyGoal:    s.DailyGoal(),
		timezone:     s.Timezone(),
		version:      s.Version(),
	}
}

func Test_UpdateUserSettingCommand_shouldSaveOnlyChangedField_whenSettingExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		call     updateCall
		expected settingValues
	}{
		{"UpdateLanguage", updateLanguage("ja"),
			settingValues{fixtureMaxWorkbooks, "ja", fixtureDailyGoal, fixtureTimezone, fixtureVersion}},
		{"UpdateDailyGoal", updateDailyGoal(40),
			settingValues{fixtureMaxWorkbooks, fixtureLanguage, 40, fixtureTimezone, fixtureVersion}},
		{"UpdateTimezone", updateTimezone("Europe/London"),
			settingValues{fixtureMaxWorkbooks, fixtureLanguage, fixtureDailyGoal, "Europe/London", fixtureVersion}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			repo := newMockuserSettingFinderSaver(t)
			repo.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(newExistingSetting(t), nil)
			var saved *domain.UserSetting
			repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, s *domain.UserSetting) error {
				saved = s
				return nil
			})
			cmd := usersetting.NewUpdateUserSettingCommand(repo)

			// when
			err := tt.call(ctx, cmd)

			// then
			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.Equal(t, tt.expected, valuesOf(saved))
		})
	}
}

func Test_UpdateUserSettingCommand_shouldSaveDefaultsWithChangedField_whenSettingNotFound(t *testing.T) {
	t.Parallel()

	defaults, err := domain.NewDefaultUserSetting(fixtureAppUserID)
	require.NoError(t, err)
	defaultMaxWorkbooks := defaults.MaxWorkbooks()

	tests := []struct {
		name     string
		call     updateCall
		expected settingValues
	}{
		{"UpdateLanguage", updateLanguage("ja"),
			settingValues{defaultMaxWorkbooks, "ja", domain.DefaultDailyGoal(), domain.DefaultTimezone(), 0}},
		{"UpdateDailyGoal", updateDailyGoal(40),
			settingValues{defaultMaxWorkbooks, domain.DefaultLanguage(), 40, domain.DefaultTimezone(), 0}},
		{"UpdateTimezone", updateTimezone("Europe/London"),
			settingValues{defaultMaxWorkbooks, domain.DefaultLanguage(), domain.DefaultDailyGoal(), "Europe/London", 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			repo := newMockuserSettingFinderSaver(t)
			repo.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, domain.ErrUserSettingNotFound)
			var saved *domain.UserSetting
			repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, s *domain.UserSetting) error {
				saved = s
				return nil
			})
			cmd := usersetting.NewUpdateUserSettingCommand(repo)

			// when
			err := tt.call(ctx, cmd)

			// then
			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.Equal(t, tt.expected, valuesOf(saved))
		})
	}
}

func Test_UpdateUserSettingCommand_shouldReturnErrInvalidValue_whenValueRejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call updateCall
	}{
		{"UpdateLanguage", updateLanguage("JA")},
		{"UpdateDailyGoal", updateDailyGoal(0)},
		{"UpdateTimezone", updateTimezone("Not/AZone")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			repo := newMockuserSettingFinderSaver(t)
			repo.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(newExistingSetting(t), nil)
			cmd := usersetting.NewUpdateUserSettingCommand(repo)

			// when
			err := tt.call(ctx, cmd)

			// then
			require.ErrorIs(t, err, usersettingservice.ErrInvalidValue)
		})
	}
}

func Test_UpdateUserSettingCommand_shouldReturnFindError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	errFind := errors.New("db connection lost")
	repo := newMockuserSettingFinderSaver(t)
	repo.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, errFind)
	cmd := usersetting.NewUpdateUserSettingCommand(repo)

	// when
	err := updateLanguage("ja")(ctx, cmd)

	// then
	require.ErrorIs(t, err, errFind)
}

func Test_UpdateUserSettingCommand_shouldNotReturnErrInvalidValue_whenStoredSettingIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: the repository fails to reconstruct a corrupted row
	repo := newMockuserSettingFinderSaver(t)
	repo.On("FindByAppUserID", mock.Anything, fixtureAppUserID).
		Return(nil, fmt.Errorf("reconstruct user setting: %w", domain.ErrInvalidArgument))
	cmd := usersetting.NewUpdateUserSettingCommand(repo)

	// when
	err := updateLanguage("ja")(ctx, cmd)

	// then
	assert.NotErrorIs(t, err, usersettingservice.ErrInvalidValue)
}

func Test_UpdateUserSettingCommand_shouldReturnSaveError_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := newMockuserSettingFinderSaver(t)
	repo.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(newExistingSetting(t), nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)
	cmd := usersetting.NewUpdateUserSettingCommand(repo)

	// when
	err := updateLanguage("ja")(ctx, cmd)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}
