// Package main is the entry point for cocotola-app.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	authinit "github.com/mocoarow/cocotola-1.26/cocotola-auth/initialize"
	questioninit "github.com/mocoarow/cocotola-1.26/cocotola-question/initialize"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"

	"github.com/mocoarow/cocotola-1.26/cocotola-app/config"
)

const appName = "cocotola-app"

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

	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, appName+"-main"))

	// init log
	shutdownLog, err := libgateway.InitLog(ctx, cfg.Log, appName)
	if err != nil {
		return 1, fmt.Errorf("init log: %w", err)
	}
	defer shutdownLog()

	// init tracer
	shutdownTrace, err := libgateway.InitTracerProvider(ctx, cfg.Trace, appName)
	if err != nil {
		return 1, fmt.Errorf("init trace: %w", err)
	}
	defer shutdownTrace()

	// init db
	dbConn, shutdownDB, err := libgateway.InitDB(ctx, cfg.DB, cfg.Log, appName)
	if err != nil {
		return 1, fmt.Errorf("init db: %w", err)
	}
	defer shutdownDB()

	// init gin
	router, err := libhandler.InitRootRouterGroup(ctx, cfg.Server, appName)
	if err != nil {
		return 1, fmt.Errorf("init router: %w", err)
	}

	// initialize auth module
	authResult, err := authinit.Initialize(ctx, router, dbConn.DB, cfg.App.Auth, cfg.App.Internal)
	if err != nil {
		return 1, fmt.Errorf("initialize auth: %w", err)
	}
	defer authResult.Close()

	// initialize question module
	orgResolver := func(ctx context.Context, name string) (int, error) {
		_, err := authResult.OrgFinder.FindByName(ctx, name)
		if err != nil {
			return 0, fmt.Errorf("find organization by name %s: %w", name, err)
		}
		// TODO(uuidv7-phase2): cocotola-question's resolver still expects int.
		// Once question migrates to UUIDs, return the VO string instead of -1.
		return -1, nil
	}
	authzAdapter := &authorizationCheckerAdapter{inner: authResult.AuthzChecker}
	questionCleanup, err := questioninit.Initialize(ctx, authResult.V1RouterGroup, cfg.App.Question, authResult.AuthMiddleware, authzAdapter, orgResolver)
	if err != nil {
		return 1, fmt.Errorf("initialize question: %w", err)
	}
	defer questionCleanup()

	// run
	readHeaderTimeout := time.Duration(cfg.Server.ReadHeaderTimeoutSec) * time.Second
	shutdownTime := time.Duration(cfg.Server.Shutdown.ShutdownTimeSec) * time.Second
	result := libprocess.Run(ctx,
		libcontroller.WithWebServerProcess(router, cfg.Server.HTTPPort, readHeaderTimeout, shutdownTime),
		libcontroller.WithMetricsServerProcess(cfg.Server.MetricsPort, readHeaderTimeout, shutdownTime),
		libgateway.WithSignalWatchProcess(),
		authResult.EventBusStart,
	)

	gracefulShutdownTime := time.Duration(cfg.Server.Shutdown.GracePeriodSec) * time.Second
	time.Sleep(gracefulShutdownTime)
	logger.InfoContext(ctx, "exited")

	return result, nil
}
