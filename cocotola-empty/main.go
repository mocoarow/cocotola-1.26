// Package main provides the cocotola-empty skeleton microservice.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

const (
	appName = "cocotola-empty"

	defaultHTTPPort             = 8080
	defaultMetricsPort          = 8081
	defaultReadHeaderTimeoutSec = 30
)

func main() {
	exitCode, err := run()
	if err != nil {
		slog.Error("run", slog.Any("error", err))
	}
	os.Exit(exitCode)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}

	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.ParseBool(value); err == nil {
			return v
		}
	}

	return defaultValue
}

func newServerConfig() libcontroller.ServerConfig {
	return libcontroller.ServerConfig{
		HTTPPort:             getEnvIntOrDefault("HTTP_PORT", defaultHTTPPort),
		MetricsPort:          getEnvIntOrDefault("METRICS_PORT", defaultMetricsPort),
		ReadHeaderTimeoutSec: getEnvIntOrDefault("READ_HEADER_TIMEOUT_SEC", defaultReadHeaderTimeoutSec),
		CORS: libcontroller.CORSConfig{
			AllowOrigins:     getEnvOrDefault("CORS_ALLOW_ORIGINS", "*"),
			AllowMethods:     getEnvOrDefault("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowHeaders:     getEnvOrDefault("CORS_ALLOW_HEADERS", "Content-Type"),
			AllowCredentials: getEnvBoolOrDefault("CORS_ALLOW_CREDENTIALS", false),
		},
		Log: libcontroller.LogConfig{
			AccessLog:             getEnvBoolOrDefault("GIN_LOG_ACCESS_LOG", true),
			AccessLogRequestBody:  getEnvBoolOrDefault("GIN_LOG_ACCESS_LOG_REQUEST_BODY", false),
			AccessLogResponseBody: getEnvBoolOrDefault("GIN_LOG_ACCESS_LOG_RESPONSE_BODY", false),
		},
		Debug: libcontroller.DebugConfig{
			Gin:  getEnvBoolOrDefault("GIN_DEBUG_GIN", false),
			Wait: getEnvBoolOrDefault("GIN_DEBUG_WAIT", false),
		},
		Shutdown: libcontroller.ShutdownConfig{
			GracePeriodSec:  getEnvIntOrDefault("SHUTDOWN_GRACE_PERIOD_SEC", 1),
			ShutdownTimeSec: getEnvIntOrDefault("SHUTDOWN_TIME_SEC", 1),
		},
	}
}

func newLogConfig() libgateway.LogConfig {
	return libgateway.LogConfig{
		Level:    getEnvOrDefault("LOG_LEVEL", "info"),
		Platform: "",
		Levels:   nil,
		Exporter: getEnvOrDefault("LOG_EXPORTER", "none"),
		OTLP:     nil,
		Uptrace:  nil,
	}
}

func newTraceConfig() libgateway.TraceConfig {
	return libgateway.TraceConfig{
		Exporter:           getEnvOrDefault("TRACE_EXPORTER", "none"),
		SamplingPercentage: 0,
		OTLP:               nil,
		Google:             nil,
		Uptrace:            nil,
	}
}

func run() (int, error) {
	ctx := context.Background()
	serverCfg := newServerConfig()
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, appName+"-main"))

	// init log
	shutdownLog, err := libgateway.InitLog(ctx, newLogConfig(), appName)
	if err != nil {
		return 1, fmt.Errorf("init log: %w", err)
	}
	defer shutdownLog()

	// init tracer
	shutdownTrace, err := libgateway.InitTracerProvider(ctx, newTraceConfig(), appName)
	if err != nil {
		return 1, fmt.Errorf("init trace: %w", err)
	}
	defer shutdownTrace()

	// init handler
	router, err := libhandler.InitRootRouterGroup(ctx, serverCfg, appName)
	if err != nil {
		return 1, fmt.Errorf("init router: %w", err)
	}

	// api
	api := router.Group("api")
	v1 := api.Group("v1")

	// public router
	test := v1.Group("test")
	test.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	test.POST("/200", func(c *gin.Context) {
		reqCtx := c.Request.Context()
		logger.InfoContext(reqCtx, "POST /200")
		params := gin.H{}
		if err := c.BindJSON(&params); err != nil {
			logger.InfoContext(reqCtx, "bind error", slog.Any("error", err))
			c.Status(http.StatusBadRequest)
			return
		}

		logger.InfoContext(reqCtx, "request params", slog.Any("params", params))
		c.Status(http.StatusOK)
	})

	// run
	readHeaderTimeout := time.Duration(serverCfg.ReadHeaderTimeoutSec) * time.Second
	shutdownTime := time.Duration(serverCfg.Shutdown.ShutdownTimeSec) * time.Second
	result := libprocess.Run(ctx,
		libcontroller.WithWebServerProcess(router, serverCfg.HTTPPort, readHeaderTimeout, shutdownTime),
		libcontroller.WithMetricsServerProcess(serverCfg.MetricsPort, readHeaderTimeout, shutdownTime),
		libgateway.WithSignalWatchProcess(),
	)

	gracefulShutdownTime := time.Duration(serverCfg.Shutdown.GracePeriodSec) * time.Second
	time.Sleep(gracefulShutdownTime)
	logger.InfoContext(ctx, "exited")

	return result, nil
}
