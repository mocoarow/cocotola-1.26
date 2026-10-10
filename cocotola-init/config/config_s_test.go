//go:build small

package config_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
)

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenVarSet(t *testing.T) {
	// given
	t.Setenv("TEST_INIT_EXPAND_VAR", "hello")

	// when
	got := config.ExpandEnvWithDefaults("TEST_INIT_EXPAND_VAR:-fallback")

	// then
	assert.Equal(t, "hello", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarUnset(t *testing.T) {
	t.Parallel()

	// given: TEST_INIT_EXPAND_UNSET is not set in environment

	// when
	got := config.ExpandEnvWithDefaults("TEST_INIT_EXPAND_UNSET:-defaultval")

	// then
	assert.Equal(t, "defaultval", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarEmpty(t *testing.T) {
	// given
	t.Setenv("TEST_INIT_EXPAND_EMPTY", "")

	// when
	got := config.ExpandEnvWithDefaults("TEST_INIT_EXPAND_EMPTY:-fallback")

	// then
	assert.Equal(t, "fallback", got)
}

func Test_expandEnvWithDefaults_shouldReturnEnvValue_withoutDefault(t *testing.T) {
	// given
	t.Setenv("TEST_INIT_EXPAND_PLAIN", "plainval")

	// when
	got := config.ExpandEnvWithDefaults("TEST_INIT_EXPAND_PLAIN")

	// then
	assert.Equal(t, "plainval", got)
}

func Test_ParseMode_shouldReturnModeInit_whenArgIsEmptyOrInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "no args", args: nil},
		{name: "init", args: []string{"init"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			got, err := config.ParseMode(tt.args)

			// then
			require.NoError(t, err)
			assert.Equal(t, config.ModeInit, got)
		})
	}
}

func Test_ParseMode_shouldReturnModeImport_whenArgIsImport(t *testing.T) {
	t.Parallel()

	// when
	got, err := config.ParseMode([]string{"import"})

	// then
	require.NoError(t, err)
	assert.Equal(t, config.ModeImport, got)
}

func Test_ParseMode_shouldReturnErrUnknownMode_whenArgIsUnknown(t *testing.T) {
	t.Parallel()

	// when
	_, err := config.ParseMode([]string{"seed"})

	// then
	require.ErrorIs(t, err, config.ErrUnknownMode)
}

func Test_ParseMode_shouldReturnErrUnknownMode_whenExtraArgsGiven(t *testing.T) {
	t.Parallel()

	// when
	_, err := config.ParseMode([]string{"import", "--dry-run"})

	// then
	require.ErrorIs(t, err, config.ErrUnknownMode)
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("INIT_POSTGRES_USERNAME", "user")
	t.Setenv("INIT_POSTGRES_PASSWORD", "password")
	t.Setenv("OWNER_LOGIN_ID", "owner")
	t.Setenv("OWNER_PASSWORD", "password123")
	t.Setenv("TRACE_EXPORTER", "google")
	t.Setenv("TRACE_GOOGLE_PROJECT_ID", "project")
}

func Test_LoadConfig_shouldSucceed_whenOwnerIsUnset(t *testing.T) {
	// given: the import job runs without owner credentials
	setRequiredEnv(t)
	t.Setenv("OWNER_LOGIN_ID", "")
	t.Setenv("OWNER_PASSWORD", "")

	// when
	_, err := config.LoadConfig()

	// then
	require.NoError(t, err)
}

func Test_LoadConfig_shouldSucceed_whenNoneExporterHasNoProjectID(t *testing.T) {
	// given
	setRequiredEnv(t)
	t.Setenv("TRACE_EXPORTER", "none")
	t.Setenv("TRACE_GOOGLE_PROJECT_ID", "")

	// when
	_, err := config.LoadConfig()

	// then
	require.NoError(t, err)
}

func Test_LoadConfig_shouldRejectProjectID_whenGoogleExporterHasNoProjectID(t *testing.T) {
	// given
	setRequiredEnv(t)
	t.Setenv("TRACE_GOOGLE_PROJECT_ID", "")

	// when
	_, err := config.LoadConfig()

	// then
	var verrs validator.ValidationErrors
	require.ErrorAs(t, err, &verrs)
	require.Len(t, verrs, 1)
	assert.Equal(t, "Config.Trace.Google.ProjectID", verrs[0].StructNamespace())
}

func validConfig() *config.Config {
	return &config.Config{
		App:      config.InitConfig{OwnerLoginID: "owner", OwnerPassword: "password123"},
		Question: config.QuestionClientConfig{BaseURL: "http://question", APIKey: "key"},
		CSVSeed:  config.CSVSeedConfig{BucketName: "questions-output"},
	}
}

func Test_Config_ValidateForMode_shouldSucceed_whenInitHasOwner(t *testing.T) {
	t.Parallel()

	// given
	cfg := validConfig()
	cfg.CSVSeed = config.CSVSeedConfig{}

	// when
	err := cfg.ValidateForMode(config.ModeInit)

	// then
	require.NoError(t, err)
}

func Test_Config_ValidateForMode_shouldReturnErrInvalidModeConfig_whenInitHasNoOwner(t *testing.T) {
	t.Parallel()

	// given
	cfg := validConfig()
	cfg.App = config.InitConfig{}

	// when
	err := cfg.ValidateForMode(config.ModeInit)

	// then
	require.ErrorIs(t, err, config.ErrInvalidModeConfig)
}

func Test_Config_ValidateForMode_shouldSucceed_whenImportHasNoOwner(t *testing.T) {
	t.Parallel()

	// given: the import job is not given owner credentials
	cfg := validConfig()
	cfg.App = config.InitConfig{}

	// when
	err := cfg.ValidateForMode(config.ModeImport)

	// then
	require.NoError(t, err)
}

func Test_Config_ValidateForMode_shouldReturnErrInvalidModeConfig_whenImportHasNoBucket(t *testing.T) {
	t.Parallel()

	// given
	cfg := validConfig()
	cfg.CSVSeed = config.CSVSeedConfig{}

	// when
	err := cfg.ValidateForMode(config.ModeImport)

	// then
	require.ErrorIs(t, err, config.ErrInvalidModeConfig)
}

func Test_Config_ValidateForMode_shouldReturnErrInvalidModeConfig_whenInitHasNoQuestionBaseURL(t *testing.T) {
	t.Parallel()

	// given
	cfg := validConfig()
	cfg.Question = config.QuestionClientConfig{}

	// when
	err := cfg.ValidateForMode(config.ModeInit)

	// then
	require.ErrorIs(t, err, config.ErrInvalidModeConfig)
}

func Test_Config_ValidateForMode_shouldReturnErrUnknownMode_whenModeIsZero(t *testing.T) {
	t.Parallel()

	// when
	err := validConfig().ValidateForMode(config.Mode(0))

	// then
	require.ErrorIs(t, err, config.ErrUnknownMode)
}

func Test_Config_ValidateForMode_shouldReturnErrInvalidModeConfig_whenImportHasNoQuestionBaseURL(t *testing.T) {
	t.Parallel()

	// given
	cfg := validConfig()
	cfg.Question = config.QuestionClientConfig{}

	// when
	err := cfg.ValidateForMode(config.ModeImport)

	// then
	require.ErrorIs(t, err, config.ErrInvalidModeConfig)
}

func Test_expandEnvWithDefaults_shouldHandleDefaultContainingColonDash(t *testing.T) {
	t.Parallel()

	// given: SplitN(..., 2) means the default value can itself contain ":-"

	// when
	got := config.ExpandEnvWithDefaults("TEST_INIT_EXPAND_MULTI_UNSET:-localhost:-5432")

	// then: default is everything after the first ":-"
	assert.Equal(t, "localhost:-5432", got)
}
