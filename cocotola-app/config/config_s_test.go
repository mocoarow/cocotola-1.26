package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-app/config"
)

const minSecretLength = 32

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_QUESTION_API_KEY", "question-api-key")
	t.Setenv("APP_AUTH_SIGNING_KEY", strings.Repeat("s", minSecretLength))
	t.Setenv("APP_AUTH_SUPABASE_JWKS_URL", "https://example.com/jwks.json")
	t.Setenv("APP_AUTH_API_KEY", strings.Repeat("k", minSecretLength))
	t.Setenv("APP_POSTGRES_USERNAME", "user")
	t.Setenv("APP_POSTGRES_PASSWORD", "password")
}

func Test_LoadConfig_shouldUseDefaultOrgResolverCacheTTL_whenEnvIsNotSet(t *testing.T) {
	// given
	setRequiredConfigEnv(t)
	t.Setenv("APP_AUTH_CLIENT_ORG_RESOLVER_CACHE_TTL_MIN", "")

	// when
	cfg, err := config.LoadConfig()

	// then
	require.NoError(t, err)
	assert.Equal(t, 60, cfg.App.AuthClient.OrgResolverCacheTTLMin)
}

func Test_LoadConfig_shouldDisableOrgResolverCache_whenEnvIsZero(t *testing.T) {
	// given
	setRequiredConfigEnv(t)
	t.Setenv("APP_AUTH_CLIENT_ORG_RESOLVER_CACHE_TTL_MIN", "0")

	// when
	cfg, err := config.LoadConfig()

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.App.AuthClient.OrgResolverCacheTTLMin)
}
