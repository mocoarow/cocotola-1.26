// Package main is the entry point for the cocotola-init bootstrap application.
// Run without arguments (or with "init") it bootstraps the organization; run
// with "import" it imports the rows appended to the CSV workbooks.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

const appName = "cocotola-init"

func main() {
	exitCode, err := run(os.Args[1:])
	if err != nil {
		slog.Error("run", slog.Any("error", err))
	}
	os.Exit(exitCode)
}

func run(args []string) (int, error) {
	ctx := context.Background()
	mode, err := config.ParseMode(args)
	if err != nil {
		return 1, fmt.Errorf("parse mode: %w", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return 1, fmt.Errorf("load config: %w", err)
	}
	if err := cfg.ValidateForMode(mode); err != nil {
		return 1, fmt.Errorf("validate config: %w", err)
	}

	shutdownLog, err := libgateway.InitLog(ctx, cfg.Log, appName)
	if err != nil {
		return 1, fmt.Errorf("init log: %w", err)
	}
	defer shutdownLog()

	shutdownTrace, err := libgateway.InitTracerProvider(ctx, cfg.Trace, appName)
	if err != nil {
		return 1, fmt.Errorf("init trace: %w", err)
	}
	defer shutdownTrace()

	dbConn, shutdownDB, err := libgateway.InitDB(ctx, cfg.DB, cfg.Log, appName)
	if err != nil {
		return 1, fmt.Errorf("init db: %w", err)
	}
	defer shutdownDB()

	if mode == config.ModeImport {
		if err := runImport(ctx, cfg, dbConn.DB); err != nil {
			return 1, fmt.Errorf("run import mode: %w", err)
		}
		return 0, nil
	}
	if err := runInit(ctx, cfg, dbConn.DB); err != nil {
		return 1, fmt.Errorf("run init mode: %w", err)
	}
	return 0, nil
}

func newQuestionAPIClient(ctx context.Context, appEnv string, qcfg config.QuestionClientConfig) (*seed.QuestionAPIClient, error) {
	// TimeoutSec == 0 means "use the default" (see config.QuestionClientConfig).
	const defaultTimeoutSec = 10
	timeoutSec := qcfg.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = defaultTimeoutSec
	}
	httpClient, err := libgateway.NewHTTPClient(ctx, appEnv, qcfg.BaseURL, time.Duration(timeoutSec)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("new http client: %w", err)
	}
	return seed.NewQuestionAPIClient(qcfg.BaseURL, qcfg.APIKey, httpClient), nil
}
