package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	return initDBDriver(ctx, cfg.DriverName, &DialectMySQL{}, db)
}

// OpenMySQLWithDSN opens a GORM MySQL connection using a raw DSN string.
func OpenMySQLWithDSN(dsn string, logLevel slog.Level, appName string) (*gorm.DB, error) {
	return openGormDB(gormmysql.Open(dsn), logLevel, appName)
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
