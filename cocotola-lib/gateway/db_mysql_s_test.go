package gateway_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

func newMySQLConfig(params map[string]string) *gateway.MySQLConfig {
	return &gateway.MySQLConfig{
		Username: "user1",
		Password: "pass1",
		Host:     "db.example.com",
		Port:     13306,
		Database: "testdb",
		Params:   params,
	}
}

func parseMySQLDSN(t *testing.T, dsn string) *mysql.Config {
	t.Helper()
	cfg, err := mysql.ParseDSN(dsn)
	require.NoError(t, err)
	return cfg
}

// mysqlDSNQuery returns the raw DSN parameters; mysql.Config does not expose charset.
func mysqlDSNQuery(t *testing.T, dsn string) url.Values {
	t.Helper()
	i := strings.LastIndex(dsn, "?")
	require.NotEqual(t, -1, i, "dsn has no query: %s", dsn)
	q, err := url.ParseQuery(dsn[i+1:])
	require.NoError(t, err)
	return q
}

func Test_BuildMySQLDSN_shouldSetConnectionFields_whenNoParamsProvided(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(nil)

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	c := parseMySQLDSN(t, dsn)
	assert.Equal(t, "user1", c.User)
	assert.Equal(t, "pass1", c.Passwd)
	assert.Equal(t, "tcp", c.Net)
	assert.Equal(t, "db.example.com:13306", c.Addr)
	assert.Equal(t, "testdb", c.DBName)
}

func Test_BuildMySQLDSN_shouldApplyFixedSettings_whenNoParamsProvided(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(nil)

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	c := parseMySQLDSN(t, dsn)
	assert.True(t, c.ParseTime)
	assert.False(t, c.MultiStatements)
	assert.Equal(t, gateway.MySQLDefaultCollation, c.Collation)
	assert.True(t, c.AllowNativePasswords)
	assert.True(t, c.CheckConnLiveness)
	assert.Equal(t, gateway.MySQLMaxAllowedPacket, c.MaxAllowedPacket)
	assert.Equal(t, time.UTC, c.Loc)
}

func Test_BuildMySQLDSN_shouldSetDefaultCharset_whenNoParamsProvided(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(nil)

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{gateway.MySQLDefaultCharset}, mysqlDSNQuery(t, dsn)["charset"])
}

func Test_BuildMySQLDSN_shouldAppendParams_whenParamsProvided(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"time_zone": "'+09:00'"})

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	c := parseMySQLDSN(t, dsn)
	assert.Equal(t, "'+09:00'", c.Params["time_zone"])
}

func Test_BuildMySQLDSN_shouldOverrideCharset_whenParamsContainsCharset(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"charset": "utf8mb3", "collation": "utf8mb3_bin"})

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"utf8mb3"}, mysqlDSNQuery(t, dsn)["charset"])
}

func Test_BuildMySQLDSN_shouldWriteCollationOnce_whenParamsContainsCollation(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"collation": "utf8mb4_general_ci"})

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"utf8mb4_general_ci"}, mysqlDSNQuery(t, dsn)["collation"])
}

func Test_BuildMySQLDSN_shouldSkipEmptyParams_whenParamValueIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"time_zone": ""})

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	_, present := mysqlDSNQuery(t, dsn)["time_zone"]
	assert.False(t, present)
}

func Test_BuildMySQLDSN_shouldKeepDefaultCharset_whenCharsetParamIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"charset": ""})

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{gateway.MySQLDefaultCharset}, mysqlDSNQuery(t, dsn)["charset"])
}

func Test_BuildMySQLDSN_shouldKeepDefaultCollation_whenCollationParamIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"collation": ""})

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{gateway.MySQLDefaultCollation}, mysqlDSNQuery(t, dsn)["collation"])
}

func Test_BuildMySQLDSN_shouldBracketIPv6Host_whenHostIsIPv6Address(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(nil)
	cfg.Host = "::1"

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	c := parseMySQLDSN(t, dsn)
	assert.Equal(t, "[::1]:13306", c.Addr)
}

func Test_BuildMySQLDSN_shouldKeepReservedCharsInPassword_whenPasswordHasReservedChars(t *testing.T) {
	t.Parallel()

	// given
	reservedChars := "p@ss/w:o?r#d"
	cfg := newMySQLConfig(nil)
	cfg.Password = reservedChars

	// when
	dsn, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.NoError(t, err)
	c := parseMySQLDSN(t, dsn)
	assert.Equal(t, reservedChars, c.Passwd)
}

func Test_BuildMySQLDSN_shouldReturnError_whenParamsContainsInvalidValue(t *testing.T) {
	t.Parallel()

	// given
	cfg := newMySQLConfig(map[string]string{"parseTime": "maybe"})

	// when
	_, err := gateway.BuildMySQLDSN(cfg)

	// then
	require.ErrorIs(t, err, gateway.ErrInvalidMySQLDSN)
}
