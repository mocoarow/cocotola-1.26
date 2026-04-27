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
	setting, err := domain.NewUserSetting(appUserID, 5, "ja")

	// then
	require.NoError(t, err)
	assert.Equal(t, appUserID, setting.AppUserID())
	assert.Equal(t, 5, setting.MaxWorkbooks())
	assert.Equal(t, "ja", setting.Language())
	assert.Equal(t, 0, setting.Version())
}

func Test_NewUserSetting_shouldReturnError_whenAppUserIDIsZero(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewUserSetting(domain.AppUserID{}, 5, "ja")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenMaxWorkbooksIsZero(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, 0, "ja")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenMaxWorkbooksIsNegative(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, -1, "ja")

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
			_, err := domain.NewUserSetting(appUserID, 5, tt.language)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
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

func Test_DefaultLanguage_shouldReturnEn(t *testing.T) {
	t.Parallel()

	// when
	lang := domain.DefaultLanguage()

	// then
	assert.Equal(t, "en", lang)
}

func Test_ReconstructUserSetting_shouldReturnUserSetting_whenValid(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.ReconstructUserSetting(appUserID, 10, "en")

	// then
	require.NoError(t, err)
	assert.Equal(t, 10, setting.MaxWorkbooks())
	assert.Equal(t, "en", setting.Language())
	assert.Equal(t, 0, setting.Version())
}

func Test_ReconstructUserSetting_shouldReturnError_whenMaxWorkbooksIsZero(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.ReconstructUserSetting(appUserID, 0, "en")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldReturnError_whenMaxWorkbooksExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	_, err := domain.NewUserSetting(appUserID, 101, "ja")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewUserSetting_shouldSucceed_whenMaxWorkbooksIsAtLimit(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")

	// when
	setting, err := domain.NewUserSetting(appUserID, 100, "ja")

	// then
	require.NoError(t, err)
	assert.Equal(t, 100, setting.MaxWorkbooks())
}

func Test_UserSetting_SetVersion_shouldSetVersion(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, _ := domain.NewUserSetting(appUserID, 5, "ja")

	// when
	setting.SetVersion(3)

	// then
	assert.Equal(t, 3, setting.Version())
}

func Test_UserSetting_ChangeLanguage_shouldUpdateLanguage(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, _ := domain.NewUserSetting(appUserID, 5, "ja")

	// when
	err := setting.ChangeLanguage("en")

	// then
	require.NoError(t, err)
	assert.Equal(t, "en", setting.Language())
}

func Test_UserSetting_ChangeLanguage_shouldReturnError_whenLanguageIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	appUserID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	setting, _ := domain.NewUserSetting(appUserID, 5, "ja")

	// when
	err := setting.ChangeLanguage("INVALID")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
