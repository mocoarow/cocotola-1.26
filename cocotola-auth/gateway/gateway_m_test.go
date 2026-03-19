//go:build medium

package gateway_test

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"testing"

	"gorm.io/gorm"

	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

var testDB *gorm.DB

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMain(m *testing.M) {
	host := getEnv("TEST_MYSQL_HOST", "127.0.0.1")
	portStr := getEnv("TEST_MYSQL_PORT", "3307")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("invalid TEST_MYSQL_PORT: %v", err)
	}
	username := getEnv("TEST_MYSQL_USERNAME", "username")
	password := getEnv("TEST_MYSQL_PASSWORD", "password")
	database := getEnv("TEST_MYSQL_DATABASE", "test")
	logLevelStr := getEnv("TEST_LOG_LEVEL", "INFO")
	logLevel := slog.LevelInfo
	if logLevelStr == "DEBUG" {
		logLevel = slog.LevelDebug
	}
	config := libgateway.MySQLConfig{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		Database: database,
	}

	// Initialize the database connection.
	db, err := libgateway.OpenMySQL(&config, logLevel, "gateway_test")
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	testDB = db

	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("get sql.DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("close sql.DB: %v", err)
		}
	}()

	exitCode := m.Run()

	os.Exit(exitCode)
}
