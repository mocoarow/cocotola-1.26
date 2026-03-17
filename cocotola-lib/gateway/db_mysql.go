package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	sloggorm "github.com/orandin/slog-gorm"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

const mysqlMaxAllowedPacket = 64 << 20 // 64 MiB

// DialectMySQL implements DialectRDBMS for MySQL.
type DialectMySQL struct {
}

// Name returns the dialect name "mysql".
func (*DialectMySQL) Name() string {
	return "mysql"
}

// BoolDefaultValue returns "0" as MySQL's false representation.
func (*DialectMySQL) BoolDefaultValue() string {
	return "0"
}

// MySQLConfig holds MySQL connection parameters.
type MySQLConfig struct {
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Host     string `yaml:"host" validate:"required"`
	Port     int    `yaml:"port" validate:"required"`
	Database string `yaml:"database" validate:"required"`
}

func initDBMySQL(ctx context.Context, cfg DBConfig, logLevel slog.Level, appName string) (*DBConnection, *sql.DB, error) {
	if cfg.MySQL == nil {
		return nil, nil, errors.New("mysql configuration is required")
	}

	db, err := OpenMySQL(cfg.MySQL, logLevel, appName)
	if err != nil {
		return nil, nil, fmt.Errorf("open mysql: %w", err)
	}

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
		DriverName: cfg.DriverName,
		Dialect:    &DialectMySQL{},
		DB:         db,
	}
	return &dbc, sqlDB, nil
}

// OpenMySQLWithDSN opens a GORM MySQL connection using a raw DSN string.
func OpenMySQLWithDSN(dsn string, logLevel slog.Level, appName string) (*gorm.DB, error) {
	gormDialector := gormmysql.Open(dsn)

	options := []sloggorm.Option{
		sloggorm.WithHandler(slog.Default().With(slog.String(logging.LoggerNameKey, appName+"-gorm")).Handler()),
	}
	if logLevel == slog.LevelDebug {
		options = append(options, sloggorm.WithTraceAll()) // trace all messages
	}

	gormConfig := gorm.Config{ //nolint:exhaustruct
		Logger:  sloggorm.New(options...),
		NowFunc: func() time.Time { return time.Now().UTC() },
	}

	db, err := gorm.Open(gormDialector, &gormConfig)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("use tracing plugin: %w", err)
	}

	return db, nil
}

// OpenMySQL opens a GORM MySQL connection using the given config.
func OpenMySQL(cfg *MySQLConfig, logLevel slog.Level, appName string) (*gorm.DB, error) {
	c := mysql.Config{ //nolint:exhaustruct
		DBName:               cfg.Database,
		User:                 cfg.Username,
		Passwd:               cfg.Password,
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Net:                  "tcp",
		ParseTime:            true,
		MultiStatements:      false,
		Params:               map[string]string{"charset": "utf8mb4"},
		Collation:            "utf8mb4_bin",
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		MaxAllowedPacket:     mysqlMaxAllowedPacket,
		Loc:                  time.UTC,
	}

	return OpenMySQLWithDSN(c.FormatDSN(), logLevel, appName)
}
