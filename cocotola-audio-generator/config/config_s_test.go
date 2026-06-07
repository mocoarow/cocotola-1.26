package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenVarIsSet(t *testing.T) {
	t.Parallel()

	// given
	t.Setenv("TEST_VAR_SET", "hello")

	// when
	got := expandEnvWithDefaults("TEST_VAR_SET:-fallback")

	// then
	assert.Equal(t, "hello", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarIsNotSet(t *testing.T) {
	t.Parallel()

	// given: TEST_VAR_UNSET is not set in the environment

	// when
	got := expandEnvWithDefaults("TEST_VAR_UNSET:-fallback")

	// then
	assert.Equal(t, "fallback", got)
}

func Test_expandEnvWithDefaults_shouldReturnDefault_whenVarIsEmptyString(t *testing.T) {
	t.Parallel()

	// given
	t.Setenv("TEST_VAR_EMPTY", "")

	// when
	got := expandEnvWithDefaults("TEST_VAR_EMPTY:-fallback")

	// then
	assert.Equal(t, "fallback", got)
}

func Test_expandEnvWithDefaults_shouldReturnEnvValue_whenNoSeparatorAndVarIsSet(t *testing.T) {
	t.Parallel()

	// given
	t.Setenv("TEST_PLAIN_VAR", "plainval")

	// when
	got := expandEnvWithDefaults("TEST_PLAIN_VAR")

	// then
	assert.Equal(t, "plainval", got)
}

func Test_expandEnvWithDefaults_shouldReturnEmpty_whenNoSeparatorAndVarIsUnset(t *testing.T) {
	t.Parallel()

	// given: TEST_PLAIN_UNSET is not set in the environment

	// when
	got := expandEnvWithDefaults("TEST_PLAIN_UNSET")

	// then
	assert.Equal(t, "", got)
}

func Test_expandEnvWithDefaults_shouldUseFirstSeparatorOnly_whenMultipleSeparatorsPresent(t *testing.T) {
	t.Parallel()

	// given: TEST_MULTI_SEP is not set, so default is everything after first ":-"

	// when
	got := expandEnvWithDefaults("TEST_MULTI_SEP:-foo:-bar")

	// then
	assert.Equal(t, "foo:-bar", got)
}
