package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

// WithWebServerProcess returns a RunProcessFunc that starts the main HTTP server.
func WithWebServerProcess(router http.Handler, port int, readHeaderTimeout, shutdownTime time.Duration) process.RunProcessFunc {
	return func(ctx context.Context) process.RunProcess {
		return func() error {
			return WebServerProcess(ctx, router, port, readHeaderTimeout, shutdownTime)
		}
	}
}

// WebServerProcess runs the HTTP server and shuts down gracefully when the context is canceled.
func WebServerProcess(ctx context.Context, router http.Handler, port int, readHeaderTimeout, shutdownTime time.Duration) error {
	logger := slog.Default().With(slog.String(logging.LoggerNameKey, "WebServer"))

	httpServer := http.Server{ //nolint:exhaustruct
		Addr:              ":" + strconv.Itoa(port),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.InfoContext(ctx, "http server listening", slog.String("addr", httpServer.Addr))

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
