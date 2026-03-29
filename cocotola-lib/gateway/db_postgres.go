package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Host     string `yaml:"host" validate:"required"`
	Port     int    `yaml:"port" validate:"required"`
	Database string `yaml:"database" validate:"required"`
	SSLMode  string `yaml:"sslMode"`
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

// OpenPostgres opens a GORM PostgreSQL connection using the given config.
func OpenPostgres(cfg *PostgresConfig, logLevel slog.Level, appName string) (*gorm.DB, error) {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, sslMode)

	return OpenPostgresWithDSN(dsn, logLevel, appName)
}
