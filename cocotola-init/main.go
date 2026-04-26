// Package main is the entry point for the cocotola-init bootstrap application.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
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

	seeder, err := buildSeeder(ctx, cfg.Question)
	if err != nil {
		return 1, fmt.Errorf("build seeder: %w", err)
	}

	slog.InfoContext(ctx, "starting initialization", slog.String("app", appName))

	if err := initialize.Initialize(ctx, dbConn.DB, seeder, cfg.App.OwnerLoginID, cfg.App.OwnerPassword); err != nil {
		return 1, fmt.Errorf("initialize: %w", err)
	}

	slog.InfoContext(ctx, "initialization completed successfully")
	return 0, nil
}

// ErrQuestionBaseURLRequired is returned by buildSeeder when the question
// client is not configured. cocotola-init refuses to start in that case
// because seeding the public space is part of its mandatory bootstrap.
var ErrQuestionBaseURLRequired = errors.New("question.baseUrl is required")

// buildSeeder constructs the public workbook seeder from the validated config.
// An empty BaseURL is treated as a configuration error rather than a silent
// skip, so misconfigured deployments fail loudly instead of leaving the
// public space empty.
func buildSeeder(ctx context.Context, qcfg config.QuestionClientConfig) (*seed.WorkbookSeeder, error) {
	if qcfg.BaseURL == "" {
		return nil, ErrQuestionBaseURLRequired
	}

	// TimeoutSec == 0 means "use the default" (see config.QuestionClientConfig).
	const defaultTimeoutSec = 10
	timeoutSec := qcfg.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = defaultTimeoutSec
	}
	timeout := time.Duration(timeoutSec) * time.Second
	httpClient, err := libgateway.NewHTTPClient(ctx, "local", qcfg.BaseURL, timeout)
	if err != nil {
		return nil, fmt.Errorf("new http client: %w", err)
	}

	seeds, err := seed.DefaultSeeds()
	if err != nil {
		return nil, fmt.Errorf("load default seeds: %w", err)
	}

	client := seed.NewQuestionAPIClient(qcfg.BaseURL, qcfg.APIKey, httpClient)
	return seed.NewWorkbookSeeder(client, seeds), nil
}
