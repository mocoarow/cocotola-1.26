// Package main is the entry point for the cocotola-init bootstrap application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

const appName = "cocotola-init"

func main() {
	exitCode, err := run()
	if err != nil {
		slog.Error("run", slog.Any("error", err))
	}
	os.Exit(exitCode)
}

func run() (int, error) {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		return 1, fmt.Errorf("load config: %w", err)
	}

	// init log
	shutdownLog, err := libgateway.InitLog(ctx, cfg.Log, appName)
	if err != nil {
		return 1, fmt.Errorf("init log: %w", err)
	}
	defer shutdownLog()

	// init db
	dbConn, shutdownDB, err := libgateway.InitDB(ctx, cfg.DB, cfg.Log, appName)
	if err != nil {
		return 1, fmt.Errorf("init db: %w", err)
	}
	defer shutdownDB()

	slog.InfoContext(ctx, "starting initialization", slog.String("app", appName))

	if err := initialize.Initialize(ctx, dbConn.DB, cfg.App.OwnerLoginID, cfg.App.OwnerPassword); err != nil {
		return 1, fmt.Errorf("initialize: %w", err)
	}

	slog.InfoContext(ctx, "initialization completed successfully")
	return 0, nil
}
