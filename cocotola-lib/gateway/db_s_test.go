package gateway_test

import (
	"context"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

func Test_DialectMySQL_Name_shouldReturnMySQL(t *testing.T) {
	t.Parallel()

	// given
	dialect := &gateway.DialectMySQL{}

	// when
	name := dialect.Name()

	// then
	if name != "mysql" {
		t.Fatalf("expected 'mysql', got %q", name)
	}
}

func Test_DialectMySQL_BoolDefaultValue_shouldReturnZero(t *testing.T) {
	t.Parallel()

	// given
	dialect := &gateway.DialectMySQL{}

	// when
	value := dialect.BoolDefaultValue()

	// then
	if value != "0" {
		t.Fatalf("expected '0', got %q", value)
	}
}

func Test_DialectPostgres_Name_shouldReturnPostgres(t *testing.T) {
	t.Parallel()

	// given
	dialect := &gateway.DialectPostgres{}

	// when
	name := dialect.Name()

	// then
	if name != "postgres" {
		t.Fatalf("expected 'postgres', got %q", name)
	}
}

func Test_DialectPostgres_BoolDefaultValue_shouldReturnFalse(t *testing.T) {
	t.Parallel()

	// given
	dialect := &gateway.DialectPostgres{}

	// when
	value := dialect.BoolDefaultValue()

	// then
	if value != "false" {
		t.Fatalf("expected 'false', got %q", value)
	}
}

func Test_InitDB_shouldReturnError_whenDriverNameIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	dbConfig := gateway.DBConfig{
		DriverName: "invalid_driver",
		MySQL:      nil,
		Postgres:   nil,
	}
	logConfig := gateway.LogConfig{
		Level:    "warn",
		Exporter: "none",
		Levels:   map[string]string{},
	}

	// when
	_, _, err := gateway.InitDB(context.Background(), dbConfig, logConfig, "test-app")

	// then
	if err == nil {
		t.Fatal("expected error for invalid driver name, got nil")
	}
}
