package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// DriverNameMySQL is the driver name for MySQL databases.
const DriverNameMySQL = "mysql"

// DBConfig holds the database driver name and driver-specific configuration.
type DBConfig struct {
	DriverName string       `yaml:"driverName" validate:"required"`
	MySQL      *MySQLConfig `yaml:"mysql"`
}

// DBConnection wraps an active GORM database connection along with its dialect.
type DBConnection struct {
	DriverName string
	Dialect    DialectRDBMS
	DB         *gorm.DB
}

func initDB(ctx context.Context, dbConfig DBConfig, logLevel slog.Level, appName string) (*DBConnection, *sql.DB, error) {
	switch dbConfig.DriverName {
	case DriverNameMySQL:
		return initDBMySQL(ctx, dbConfig, logLevel, appName)
	default:
		return nil, nil, fmt.Errorf("invalid database driver: %s", dbConfig.DriverName)
	}
}

// InitDB initializes a database connection based on the configured driver.
// It returns the connection and a cleanup function to close the underlying sql.DB.
func InitDB(ctx context.Context, dbConfig DBConfig, logConfig LogConfig, appName string) (*DBConnection, func(), error) {
	dbLogLevel := slog.LevelWarn
	if level, ok := logConfig.Levels["db"]; ok {
		dbLogLevel = stringToLogLevel(level)
	}

	dbc, sqlDB, err := initDB(ctx, dbConfig, dbLogLevel, appName)
	if err != nil {
		return nil, nil, fmt.Errorf("init DB: %w", err)
	}

	return dbc, func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("close sqlDB", "error", err)
		}
	}, nil
}

// DialectRDBMS abstracts database dialect differences (e.g. default values).
type DialectRDBMS interface {
	Name() string
	BoolDefaultValue() string
}
