package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DialectPostgres implements DialectRDBMS for PostgreSQL.
type DialectPostgres struct {
}

// Name returns the dialect name "postgres".
func (*DialectPostgres) Name() string {
	return "postgres"
}

// BoolDefaultValue returns "false" as PostgreSQL's false representation.
func (*DialectPostgres) BoolDefaultValue() string {
	return "false"
}

// PostgresConfig holds PostgreSQL connection parameters.
type PostgresConfig struct {
	Username string            `yaml:"username" validate:"required"`
	Password string            `yaml:"password" validate:"required"`
	Host     string            `yaml:"host" validate:"required"`
	Port     int               `yaml:"port" validate:"required"`
	Database string            `yaml:"database" validate:"required"`
	SSLMode  string            `yaml:"sslMode"`
	// Params are extra connection parameters appended to the DSN. They are
	// applied after the struct-level fields, so a key of "sslmode" or
	// "TimeZone" here overrides the value derived from SSLMode or the default
	// UTC time zone.
	Params map[string]string `yaml:"params"`
}

func initDBPostgres(ctx context.Context, cfg DBConfig, logLevel slog.Level, appName string) (*DBConnection, *sql.DB, error) {
	if cfg.Postgres == nil {
		return nil, nil, errors.New("postgres configuration is required")
	}

	db, err := OpenPostgres(cfg.Postgres, logLevel, appName)
	if err != nil {
		return nil, nil, fmt.Errorf("open postgres: %w", err)
	}

	return initDBDriver(ctx, cfg.DriverName, &DialectPostgres{}, db)
}

// OpenPostgresWithDSN opens a GORM PostgreSQL connection using a raw DSN string.
func OpenPostgresWithDSN(dsn string, logLevel slog.Level, appName string) (*gorm.DB, error) {
	return openGormDB(gormpostgres.Open(dsn), logLevel, appName)
}

// BuildPostgresDSN builds a PostgreSQL DSN string from the given config.
// The DSN uses URL form so userinfo, host, and query params are escaped by
// net/url, and pgconn.ParseConfig validates the result rejecting malformed
// connection strings before they reach the driver.
func BuildPostgresDSN(cfg *PostgresConfig) (string, error) {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	q := url.Values{}
	q.Set("sslmode", sslMode)
	q.Set("TimeZone", "UTC")
	for k, v := range cfg.Params {
		if v != "" {
			q.Set(k, v)
		}
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     "/" + cfg.Database,
		RawQuery: q.Encode(),
	}
	dsn := u.String()

	if _, err := pgconn.ParseConfig(dsn); err != nil {
		return "", fmt.Errorf("parse postgres dsn: %w", err)
	}
	return dsn, nil
}

// OpenPostgres opens a GORM PostgreSQL connection using the given config.
func OpenPostgres(cfg *PostgresConfig, logLevel slog.Level, appName string) (*gorm.DB, error) {
	dsn, err := BuildPostgresDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("build postgres dsn: %w", err)
	}
	return OpenPostgresWithDSN(dsn, logLevel, appName)
}
