// Package main is the entry point for cocotola-app.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	authinit "github.com/mocoarow/cocotola-1.26/cocotola-auth/initialize"
	questiongateway "github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
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
	authResult, err := authinit.Initialize(ctx, router, dbConn.DB, cfg.App.Auth)
	if err != nil {
		return 1, fmt.Errorf("initialize auth: %w", err)
	}
	defer authResult.Close()

	// auth HTTP client for question module to call auth internal APIs.
	// In monolith mode, this calls localhost (same process). This allows the question
	// module to use the same HTTP-based interface as in standalone microservice mode,
	// enabling seamless deployment as either a monolith or separate services.
	authClientBaseURL := cfg.App.AuthClient.BaseURL
	authClientAPIKey := cfg.App.Auth.APIKey
	authClientTimeout := time.Duration(cfg.App.AuthClient.TimeoutSec) * time.Second

	httpClient, err := libgateway.NewHTTPClient(ctx, "local", authClientBaseURL, authClientTimeout)
	if err != nil {
		return 1, fmt.Errorf("create auth HTTP client: %w", err)
	}

	authMiddleware := questiongateway.NewAuthMiddleware(authClientBaseURL, httpClient)
	authzChecker := questiongateway.NewAuthServiceAuthorizationChecker(authClientBaseURL, authClientAPIKey, httpClient)
	orgResolver := questiongateway.AuthServiceOrganizationResolver(authClientBaseURL, authClientAPIKey, httpClient)
	maxWbFetcher := questiongateway.NewAuthServiceMaxWorkbooksFetcher(authClientBaseURL, authClientAPIKey, httpClient)
	policyAdder := questiongateway.NewAuthServicePolicyAdder(authClientBaseURL, authClientAPIKey, httpClient)

	// initialize question module
	questionCleanup, err := questioninit.Initialize(ctx, authResult.V1RouterGroup, cfg.App.Question, authMiddleware, authzChecker, orgResolver, maxWbFetcher, policyAdder)
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
