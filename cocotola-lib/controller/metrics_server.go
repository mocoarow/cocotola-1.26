package controller

import (
	"context"
	"errors"
	"fmt"
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

	errCh := make(chan error)

	go func() {
		defer close(errCh)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "listen and serve", slog.Any("error", err))

			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTime)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "server forced to shutdown", slog.Any("error", err))

			return fmt.Errorf("httpServer.Shutdown: %w", err)
		}

		return nil
	case err := <-errCh:
		return fmt.Errorf("httpServer.ListenAndServe: %w", err)
	}
}
