package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

// WithWebServerProcess returns a RunProcessFunc that starts the main HTTP server.
func WithWebServerProcess(router http.Handler, port int, timeouts HTTPServerTimeouts) process.RunProcessFunc {
	return func(ctx context.Context) process.RunProcess {
		return func() error {
			return WebServerProcess(ctx, router, port, timeouts)
		}
	}
}

// WebServerProcess runs the HTTP server and shuts down gracefully when the context is canceled.
func WebServerProcess(ctx context.Context, router http.Handler, port int, timeouts HTTPServerTimeouts) error {
	logger := slog.Default().With(slog.String(logging.LoggerNameKey, "WebServer"))

	httpServer := newHTTPServer(port, router, timeouts)

	logger.InfoContext(ctx, "http server listening", slog.String("addr", httpServer.Addr))

	return runHTTPServer(ctx, httpServer, logger, timeouts.Shutdown)
}
