package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	sloggorm "github.com/orandin/slog-gorm"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// DriverNameMySQL is the driver name for MySQL databases.
const DriverNameMySQL = "mysql"

// DriverNamePostgres is the driver name for PostgreSQL databases.
const DriverNamePostgres = "postgres"

// DBConfig holds the database driver name and driver-specific configuration.
type DBConfig struct {
	DriverName string          `yaml:"driverName" validate:"required"`
	MySQL      *MySQLConfig    `yaml:"mysql"`
	Postgres   *PostgresConfig `yaml:"postgres"`
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
	case DriverNamePostgres:
		return initDBPostgres(ctx, dbConfig, logLevel, appName)
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

// openGormDB opens a GORM connection with the given dialector and common configuration.
func openGormDB(dialector gorm.Dialector, logLevel slog.Level, appName string) (*gorm.DB, error) {
	options := []sloggorm.Option{
		sloggorm.WithHandler(slog.Default().With(slog.String(logging.LoggerNameKey, appName+"-gorm")).Handler()),
	}
	if logLevel == slog.LevelDebug {
		options = append(options, sloggorm.WithTraceAll()) // trace all messages
	}

	gormConfig := gorm.Config{ //nolint:exhaustruct
		Logger:         sloggorm.New(options...),
		NowFunc:        func() time.Time { return time.Now().UTC() },
		TranslateError: true,
	}

	db, err := gorm.Open(dialector, &gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("use tracing plugin: %w", err)
	}

	return db, nil
}

// initDBDriver opens a GORM connection, pings it, and wraps it in a DBConnection.
func initDBDriver(ctx context.Context, driverName string, dialect DialectRDBMS, db *gorm.DB) (*DBConnection, *sql.DB, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get db: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("close sqlDB after ping failure", slog.Any("error", closeErr))
		}

		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	dbc := DBConnection{
		DriverName: driverName,
		Dialect:    dialect,
		DB:         db,
	}
	return &dbc, sqlDB, nil
}
