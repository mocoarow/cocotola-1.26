package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/config"
)

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenVarIsSet(t *testing.T) {
	// given
	t.Setenv("TEST_EXPAND_VAR", "hello")

	// when
	got := config.ExpandEnvWithDefaults("TEST_EXPAND_VAR:-fallback")

	// then
	assert.Equal(t, "hello", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarIsUnset(t *testing.T) {
	t.Parallel()

	// given, when, then
	got := config.ExpandEnvWithDefaults("TEST_EXPAND_UNSET:-defaultval")

	assert.Equal(t, "defaultval", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarIsEmpty(t *testing.T) {
	// given
	t.Setenv("TEST_EXPAND_EMPTY", "")

	// when
	got := config.ExpandEnvWithDefaults("TEST_EXPAND_EMPTY:-fallback")

	// then
	assert.Equal(t, "fallback", got)
}

func Test_expandEnvWithDefaults_shouldReturnFirstDefault_whenMultipleSeparators(t *testing.T) {
	t.Parallel()

	// given, when, then
	got := config.ExpandEnvWithDefaults("TEST_EXPAND_MULTI_UNSET:-first:-second")

	assert.Equal(t, "first:-second", got)
}

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenNoSeparator(t *testing.T) {
	// given
	t.Setenv("TEST_EXPAND_PLAIN", "plainval")

	// when
	got := config.ExpandEnvWithDefaults("TEST_EXPAND_PLAIN")

	// then
	assert.Equal(t, "plainval", got)
}

func Test_expandEnvWithDefaults_shouldReturnEmpty_whenNoSeparatorAndVarUnset(t *testing.T) {
	t.Parallel()

	// given, when, then
	got := config.ExpandEnvWithDefaults("TEST_EXPAND_NODEFAULT_UNSET")

	assert.Empty(t, got)
}
