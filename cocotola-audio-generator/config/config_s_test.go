package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/config"
)

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenVariableIsSet(t *testing.T) {
	t.Setenv("TEST_VAR_155", "actual_value")

	// when
	got := config.ExpandEnvWithDefaults("TEST_VAR_155:-default_value")

	// then
	assert.Equal(t, "actual_value", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVariableIsNotSet(t *testing.T) {
	// given: TEST_VAR_155_UNSET is not set in the environment

	// when
	got := config.ExpandEnvWithDefaults("TEST_VAR_155_UNSET:-default_value")

	// then
	assert.Equal(t, "default_value", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVariableIsEmptyString(t *testing.T) {
	t.Setenv("TEST_VAR_155_EMPTY", "")

	// when
	got := config.ExpandEnvWithDefaults("TEST_VAR_155_EMPTY:-default_value")

	// then
	assert.Equal(t, "default_value", got)
}

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenNoSeparator(t *testing.T) {
	t.Setenv("TEST_VAR_155_PLAIN", "plain_value")

	// when
	got := config.ExpandEnvWithDefaults("TEST_VAR_155_PLAIN")

	// then
	assert.Equal(t, "plain_value", got)
}

func Test_expandEnvWithDefaults_shouldReturnEmpty_whenNoSeparatorAndVarNotSet(t *testing.T) {
	// given: TEST_VAR_155_PLAIN_UNSET is not set in the environment

	// when
	got := config.ExpandEnvWithDefaults("TEST_VAR_155_PLAIN_UNSET")

	// then
	assert.Equal(t, "", got)
}

func Test_expandEnvWithDefaults_shouldUseFirstSeparatorOnly_whenMultipleSeparatorsPresent(t *testing.T) {
	// given: TEST_VAR_155_MULTI is not set; default contains ":-" itself
	// strings.Cut splits at the first ":-" so the default becomes "a:-b"

	// when
	got := config.ExpandEnvWithDefaults("TEST_VAR_155_MULTI:-a:-b")

	// then
	assert.Equal(t, "a:-b", got)
}
