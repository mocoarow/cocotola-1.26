//go:build small

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

func Test_expandEnvWithDefaults_shouldHandleDefaultContainingColonDash(t *testing.T) {
	t.Parallel()

	// given: SplitN(..., 2) means the default value can itself contain ":-"

	// when
	got := config.ExpandEnvWithDefaults("TEST_INIT_EXPAND_MULTI_UNSET:-localhost:-5432")

	// then: default is everything after the first ":-"
	assert.Equal(t, "localhost:-5432", got)
}
