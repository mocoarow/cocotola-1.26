package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/config"
)

func Test_ExpandEnvWithDefaults_shouldReturnDefault_whenEnvIsNotSet(t *testing.T) {
	t.Parallel()

	// given
	varName := "COCOTOLA_TEST_UNSET_VAR:-fallback"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Equal(t, "fallback", result)
}

func Test_ExpandEnvWithDefaults_shouldReturnEnvValue_whenEnvIsSet(t *testing.T) {
	// given
	t.Setenv("COCOTOLA_TEST_SET_VAR", "from_env")
	varName := "COCOTOLA_TEST_SET_VAR:-fallback"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Equal(t, "from_env", result)
}

func Test_ExpandEnvWithDefaults_shouldReturnEmpty_whenNoDefaultAndEnvNotSet(t *testing.T) {
	t.Parallel()

	// given
	varName := "COCOTOLA_TEST_MISSING_VAR"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Empty(t, result)
}

func Test_ExpandEnvWithDefaults_shouldReturnEmptyDefault_whenDefaultIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	varName := "COCOTOLA_TEST_UNSET_VAR2:-"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Empty(t, result)
}

func Test_ExpandEnvWithDefaults_shouldReturnDefaultWithColonDash_whenDefaultContainsSeparator(t *testing.T) {
	t.Parallel()

	// given
	varName := "COCOTOLA_TEST_UNSET_VAR3:-value:-with:-separators"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Equal(t, "value:-with:-separators", result)
}

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("QUESTION_API_KEY", "question-api-key")
	t.Setenv("AUTH_API_KEY", "auth-api-key")
	t.Setenv("TRACE_GOOGLE_PROJECT_ID", "test-project")
}

func Test_LoadConfig_shouldUseDefaultOrgResolverCacheTTL_whenEnvIsNotSet(t *testing.T) {
	// given
	setRequiredConfigEnv(t)
	t.Setenv("ORG_RESOLVER_CACHE_TTL_MIN", "")

	// when
	cfg, err := config.LoadConfig()

	// then
	require.NoError(t, err)
	assert.Equal(t, 60, cfg.Auth.OrgResolverCacheTTLMin)
}

func Test_LoadConfig_shouldDisableOrgResolverCache_whenEnvIsZero(t *testing.T) {
	// given
	setRequiredConfigEnv(t)
	t.Setenv("ORG_RESOLVER_CACHE_TTL_MIN", "0")

	// when
	cfg, err := config.LoadConfig()

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Auth.OrgResolverCacheTTLMin)
}

func Test_LoadConfig_shouldUseDefaultMeCacheTTL_whenEnvIsNotSet(t *testing.T) {
	// given
	setRequiredConfigEnv(t)
	t.Setenv("AUTH_ME_CACHE_TTL_SEC", "")

	// when
	cfg, err := config.LoadConfig()

	// then
	require.NoError(t, err)
	assert.Equal(t, 90, cfg.Auth.MeCacheTTLSec)
}

func Test_LoadConfig_shouldDisableMeCache_whenEnvIsZero(t *testing.T) {
	// given
	setRequiredConfigEnv(t)
	t.Setenv("AUTH_ME_CACHE_TTL_SEC", "0")

	// when
	cfg, err := config.LoadConfig()

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Auth.MeCacheTTLSec)
}

func Test_ExpandEnvWithDefaults_shouldReturnEnvValue_whenEnvIsSetAndNoDefault(t *testing.T) {
	// given
	t.Setenv("COCOTOLA_TEST_PLAIN_VAR", "plain_value")
	varName := "COCOTOLA_TEST_PLAIN_VAR"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Equal(t, "plain_value", result)
}
