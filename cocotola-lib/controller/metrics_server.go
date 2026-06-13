package controller

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

// WithMetricsServerProcess returns a RunProcessFunc that starts the metrics/healthcheck server.
func WithMetricsServerProcess(port int, readHeaderTimeout, shutdownTime time.Duration) process.RunProcessFunc {
	return func(ctx context.Context) process.RunProcess {
		return func() error {
			return MetricsServerProcess(ctx, port, readHeaderTimeout, shutdownTime)
		}
	}
}

// MetricsServerProcess runs the metrics server exposing /healthcheck and /metrics endpoints.
func MetricsServerProcess(ctx context.Context, port int, readHeaderTimeout, shutdownTime time.Duration) error {
	logger := slog.Default().With(slog.String(logging.LoggerNameKey, "MetricsServer"))
	router := gin.New()
	router.Use(gin.Recovery())

	httpServer := http.Server{ //nolint:exhaustruct
		Addr:              ":" + strconv.Itoa(port),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	router.GET("/healthcheck", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	logger.InfoContext(ctx, "metrics server listening", slog.String("addr", httpServer.Addr))

	return runHTTPServer(ctx, &httpServer, logger, shutdownTime)
}
