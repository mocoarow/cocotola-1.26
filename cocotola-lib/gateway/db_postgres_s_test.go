package gateway_test

import (
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

// parseDSN re-parses the built DSN to extract structured fields for assertions.
func parseDSN(t *testing.T, dsn string) (*url.URL, *pgconn.Config) {
	t.Helper()
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	cfg, err := pgconn.ParseConfig(dsn)
	require.NoError(t, err)
	return u, cfg
}

func Test_BuildPostgresDSN_shouldBuildBasicDSN_whenNoParamsProvided(t *testing.T) {
	t.Parallel()

	// given
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then
	require.NoError(t, err)

	u, pg := parseDSN(t, dsn)
	assert.Equal(t, "postgres", u.Scheme)
	assert.Equal(t, "localhost:5432", u.Host)
	assert.Equal(t, "/testdb", u.Path)
	assert.Equal(t, "user1", pg.User)
	assert.Equal(t, "pass1", pg.Password)
	assert.Equal(t, "UTC", pg.RuntimeParams["TimeZone"])
}

func Test_BuildPostgresDSN_shouldDefaultSSLModeToDisable_whenSSLModeIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "",
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then
	require.NoError(t, err)
	u, _ := parseDSN(t, dsn)
	assert.Equal(t, "disable", u.Query().Get("sslmode"))
}

func Test_BuildPostgresDSN_shouldAppendParams_whenParamsProvided(t *testing.T) {
	t.Parallel()

	// given
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
		Params: map[string]string{
			"application_name": "myapp",
		},
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then
	require.NoError(t, err)
	_, pg := parseDSN(t, dsn)
	assert.Equal(t, "myapp", pg.RuntimeParams["application_name"])
}

func Test_BuildPostgresDSN_shouldSkipEmptyParams_whenParamValueIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
		Params: map[string]string{
			"application_name": "",
		},
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then
	require.NoError(t, err)
	u, _ := parseDSN(t, dsn)
	_, present := u.Query()["application_name"]
	assert.False(t, present)
}

func Test_BuildPostgresDSN_shouldEscapeSingleQuoteInParamValue_whenParamContainsSingleQuote(t *testing.T) {
	t.Parallel()

	// given
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
		Params: map[string]string{
			"application_name": "it's app",
		},
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then: pgconn parses the URL-encoded value back to the literal string
	require.NoError(t, err)
	_, pg := parseDSN(t, dsn)
	assert.Equal(t, "it's app", pg.RuntimeParams["application_name"])
}

func Test_BuildPostgresDSN_shouldEscapeBackslashInParamValue_whenParamContainsBackslash(t *testing.T) {
	t.Parallel()

	// given
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
		Params: map[string]string{
			"application_name": `back\slash`,
		},
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then
	require.NoError(t, err)
	_, pg := parseDSN(t, dsn)
	assert.Equal(t, `back\slash`, pg.RuntimeParams["application_name"])
}

func Test_BuildPostgresDSN_shouldPreserveSpaceInParamValue_whenParamContainsSpace(t *testing.T) {
	t.Parallel()

	// given: param value contains a space
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
		Params: map[string]string{
			"application_name": "my app",
		},
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then
	require.NoError(t, err)
	_, pg := parseDSN(t, dsn)
	assert.Equal(t, "my app", pg.RuntimeParams["application_name"])
}

func Test_BuildPostgresDSN_shouldBracketIPv6Host_whenHostIsIPv6Address(t *testing.T) {
	t.Parallel()

	// given: host is an IPv6 literal that must be bracketed in URL form
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "::1",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then: pgconn resolves the bracketed host back to the literal address
	require.NoError(t, err)
	u, pg := parseDSN(t, dsn)
	assert.Equal(t, "[::1]:5432", u.Host)
	assert.Equal(t, "::1", pg.Host)
}

func Test_BuildPostgresDSN_shouldEscapeSpecialCharsInPassword_whenPasswordHasReservedChars(t *testing.T) {
	t.Parallel()

	// given: password contains URL-reserved chars (@, /, :, ?, #)
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "p@ss/w:o?r#d",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then: pgconn decodes the password back to its literal form
	require.NoError(t, err)
	_, pg := parseDSN(t, dsn)
	assert.Equal(t, "p@ss/w:o?r#d", pg.Password)
}

func Test_BuildPostgresDSN_shouldEscapeSpecialCharsInUsername_whenUsernameHasReservedChars(t *testing.T) {
	t.Parallel()

	// given: username contains URL-reserved chars (@, /, :)
	cfg := &gateway.PostgresConfig{
		Username: "us@r/na:me",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then: pgconn decodes the username back to its literal form
	require.NoError(t, err)
	_, pg := parseDSN(t, dsn)
	assert.Equal(t, "us@r/na:me", pg.User)
}

func Test_BuildPostgresDSN_shouldOverrideSSLMode_whenParamsContainsSSLMode(t *testing.T) {
	t.Parallel()

	// given: a sslmode key in Params conflicts with the struct-level SSLMode
	cfg := &gateway.PostgresConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		SSLMode:  "disable",
		Params: map[string]string{
			"sslmode": "require",
		},
	}

	// when
	dsn, err := gateway.BuildPostgresDSN(cfg)

	// then: the Params value wins over the struct-level SSLMode
	require.NoError(t, err)
	u, _ := parseDSN(t, dsn)
	assert.Equal(t, "require", u.Query().Get("sslmode"))
}
