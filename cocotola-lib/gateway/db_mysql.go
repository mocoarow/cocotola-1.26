package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	mysqlMaxAllowedPacket = 64 << 20 // 64 MiB
	mysqlDefaultCharset   = "utf8mb4"
	mysqlDefaultCollation = "utf8mb4_bin"
	mysqlCharsetParam     = "charset"
	mysqlCollationParam   = "collation"
)

// ErrInvalidMySQLDSN is returned when the built DSN is rejected by the driver.
var ErrInvalidMySQLDSN = errors.New("invalid mysql dsn")

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
	// Params are passed to go-sql-driver/mysql as DSN parameters.
	Params map[string]string `yaml:"params"`
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

// BuildMySQLDSN builds a MySQL DSN string from the given config.
func BuildMySQLDSN(cfg *MySQLConfig) (string, error) {
	params := map[string]string{mysqlCharsetParam: mysqlDefaultCharset}
	collation := mysqlDefaultCollation
	for k, v := range cfg.Params {
		switch {
		case v == "":
		case k == mysqlCollationParam:
			collation = v
		default:
			params[k] = v
		}
	}

	c := mysql.NewConfig()
	c.DBName = cfg.Database
	c.User = cfg.Username
	c.Passwd = cfg.Password
	c.Addr = net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	c.Net = "tcp"
	c.ParseTime = true
	c.MultiStatements = false
	c.Params = params
	c.Collation = collation
	c.AllowNativePasswords = true
	c.CheckConnLiveness = true
	c.MaxAllowedPacket = mysqlMaxAllowedPacket
	c.Loc = time.UTC
	dsn := c.FormatDSN()

	if _, err := mysql.ParseDSN(dsn); err != nil {
		return "", fmt.Errorf("parse mysql dsn: %w: %w", ErrInvalidMySQLDSN, err)
	}
	return dsn, nil
}

// OpenMySQL opens a GORM MySQL connection using the given config.
func OpenMySQL(cfg *MySQLConfig, logLevel slog.Level, appName string) (*gorm.DB, error) {
	dsn, err := BuildMySQLDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("build mysql dsn: %w", err)
	}
	return OpenMySQLWithDSN(dsn, logLevel, appName)
}
