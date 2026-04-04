package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	assert.Equal(t, "", result)
}

func Test_ExpandEnvWithDefaults_shouldReturnEmptyDefault_whenDefaultIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	varName := "COCOTOLA_TEST_UNSET_VAR2:-"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Equal(t, "", result)
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

func Test_ExpandEnvWithDefaults_shouldReturnEnvValue_whenEnvIsSetAndNoDefault(t *testing.T) {
	// given
	t.Setenv("COCOTOLA_TEST_PLAIN_VAR", "plain_value")
	varName := "COCOTOLA_TEST_PLAIN_VAR"

	// when
	result := config.ExpandEnvWithDefaults(varName)

	// then
	assert.Equal(t, "plain_value", result)
}
