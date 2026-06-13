package controller

import (
	"context"
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

	return runHTTPServer(ctx, &httpServer, logger, shutdownTime)
}
