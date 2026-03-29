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
	host := getEnv("TEST_POSTGRES_HOST", "127.0.0.1")
	portStr := getEnv("TEST_POSTGRES_PORT", "5433")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("invalid TEST_POSTGRES_PORT: %v", err)
	}
	username := getEnv("TEST_POSTGRES_USERNAME", "username")
	password := getEnv("TEST_POSTGRES_PASSWORD", "password")
	database := getEnv("TEST_POSTGRES_DATABASE", "test")
	logLevelStr := getEnv("TEST_LOG_LEVEL", "INFO")
	logLevel := slog.LevelInfo
	if logLevelStr == "DEBUG" {
		logLevel = slog.LevelDebug
	}
	config := libgateway.PostgresConfig{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		Database: database,
		SSLMode:  "disable",
	}

	// Initialize the database connection.
	db, err := libgateway.OpenPostgres(&config, logLevel, "gateway_test")
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	testDB = db

	exitCode := m.Run()

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("get sql.DB: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close sql.DB: %v", err)
		}
	}

	os.Exit(exitCode)
}
