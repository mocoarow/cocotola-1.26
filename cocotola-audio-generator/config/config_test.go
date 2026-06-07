package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenVarIsSet(t *testing.T) {
	t.Parallel()

	// given
	t.Setenv("TEST_EXPAND_VAR", "hello")

	// when
	got := expandEnvWithDefaults("TEST_EXPAND_VAR:-fallback")

	// then
	assert.Equal(t, "hello", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarIsUnset(t *testing.T) {
	t.Parallel()

	// given: TEST_EXPAND_UNSET is not set in environment

	// when
	got := expandEnvWithDefaults("TEST_EXPAND_UNSET:-defaultval")

	// then
	assert.Equal(t, "defaultval", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	t.Setenv("TEST_EXPAND_EMPTY", "")

	// when
	got := expandEnvWithDefaults("TEST_EXPAND_EMPTY:-fallback")

	// then
	assert.Equal(t, "fallback", got)
}

func Test_expandEnvWithDefaults_shouldReturnFirstDefault_whenMultipleSeparators(t *testing.T) {
	t.Parallel()

	// given: strings.Cut stops at the first ":-", so the default itself may contain ":-"

	// when
	got := expandEnvWithDefaults("TEST_EXPAND_MULTI_UNSET:-first:-second")

	// then: default value is everything after the first ":-"
	assert.Equal(t, "first:-second", got)
}

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenNoSeparator(t *testing.T) {
	t.Parallel()

	// given
	t.Setenv("TEST_EXPAND_PLAIN", "plainval")

	// when
	got := expandEnvWithDefaults("TEST_EXPAND_PLAIN")

	// then
	assert.Equal(t, "plainval", got)
}

func Test_expandEnvWithDefaults_shouldReturnEmpty_whenNoSeparatorAndVarUnset(t *testing.T) {
	t.Parallel()

	// given: TEST_EXPAND_NODEFAULT_UNSET is not set in environment

	// when
	got := expandEnvWithDefaults("TEST_EXPAND_NODEFAULT_UNSET")

	// then
	assert.Equal(t, "", got)
}
