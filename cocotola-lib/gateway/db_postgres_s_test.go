package gateway_test

import (
	"strings"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

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
	dsn := gateway.BuildPostgresDSN(cfg)

	// then
	expected := "host=localhost user=user1 password=pass1 dbname=testdb port=5432 sslmode=disable TimeZone=UTC"
	if dsn != expected {
		t.Fatalf("expected %q, got %q", expected, dsn)
	}
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
	dsn := gateway.BuildPostgresDSN(cfg)

	// then
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Fatalf("expected sslmode=disable in DSN, got %q", dsn)
	}
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
			"default_query_exec_mode": "simple_protocol",
		},
	}

	// when
	dsn := gateway.BuildPostgresDSN(cfg)

	// then
	if !strings.Contains(dsn, "default_query_exec_mode='simple_protocol'") {
		t.Fatalf("expected default_query_exec_mode='simple_protocol' in DSN, got %q", dsn)
	}
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
			"default_query_exec_mode": "",
		},
	}

	// when
	dsn := gateway.BuildPostgresDSN(cfg)

	// then
	if strings.Contains(dsn, "default_query_exec_mode") {
		t.Fatalf("expected empty param to be skipped, got %q", dsn)
	}
}

func Test_BuildPostgresDSN_shouldQuoteParamValues_whenParamsProvided(t *testing.T) {
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
			"application_name": "my app",
		},
	}

	// when
	dsn := gateway.BuildPostgresDSN(cfg)

	// then
	if !strings.Contains(dsn, "application_name='my app'") {
		t.Fatalf("expected quoted param value in DSN, got %q", dsn)
	}
}
