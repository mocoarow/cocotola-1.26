// Package main is the entry point for the standalone cocotola-question microservice.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
	questioninit "github.com/mocoarow/cocotola-1.26/cocotola-question/initialize"
)

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

	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-main"))

	// init log
	shutdownLog, err := libgateway.InitLog(ctx, cfg.Log, domain.AppName)
	if err != nil {
		return 1, fmt.Errorf("init log: %w", err)
	}
	defer shutdownLog()

	// init tracer
	shutdownTrace, err := libgateway.InitTracerProvider(ctx, cfg.Trace, domain.AppName)
	if err != nil {
		return 1, fmt.Errorf("init trace: %w", err)
	}
	defer shutdownTrace()

	// init gin
	router, err := libhandler.InitRootRouterGroup(ctx, cfg.Server, domain.AppName)
	if err != nil {
		return 1, fmt.Errorf("init router: %w", err)
	}

	// auth HTTP client for communicating with cocotola-auth
	authTimeout := time.Duration(cfg.Auth.TimeoutSec) * time.Second
	authAudience := cfg.Auth.Audience
	if authAudience == "" {
		authAudience = cfg.Auth.BaseURL
	}

	httpClient, err := libgateway.NewHTTPClient(ctx, cfg.AppEnv, authAudience, authTimeout)
	if err != nil {
		return 1, fmt.Errorf("create auth HTTP client: %w", err)
	}

	// auth middleware (validates tokens via cocotola-auth API)
	authMiddleware := gateway.NewAuthMiddleware(cfg.Auth.BaseURL, httpClient)

	// authorization checker (checks RBAC via cocotola-auth internal API)
	authzChecker := gateway.NewAuthServiceAuthorizationChecker(cfg.Auth.BaseURL, cfg.Auth.APIKey, httpClient)

	// organization resolver (resolves org name to ID via cocotola-auth internal API)
	orgResolver := gateway.AuthServiceOrganizationResolver(cfg.Auth.BaseURL, cfg.Auth.APIKey, httpClient)

	// initialize question module
	api := router.Group("api")
	v1 := api.Group("v1")

	questionCleanup, err := questioninit.Initialize(ctx, v1, cfg.Question, authMiddleware, authzChecker, orgResolver)
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
	)

	gracefulShutdownTime := time.Duration(cfg.Server.Shutdown.GracePeriodSec) * time.Second
	time.Sleep(gracefulShutdownTime)
	logger.InfoContext(ctx, "exited")

	return result, nil
}
