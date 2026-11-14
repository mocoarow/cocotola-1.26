package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func runHTTPServer(ctx context.Context, server *http.Server, logger *slog.Logger, shutdownTime time.Duration) error {
	lc := net.ListenConfig{
		Control:   nil,
		KeepAlive: 0,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   false,
			Idle:     0,
			Interval: 0,
			Count:    0,
		},
	}
	l, err := lc.Listen(ctx, "tcp", server.Addr)
	if err != nil {
		logger.ErrorContext(ctx, "listen", slog.Any("error", err))

		return fmt.Errorf("listen %s: %w", server.Addr, err)
	}

	return serveUntilDone(ctx, server, l, logger, shutdownTime)
}

func serveUntilDone(ctx context.Context, server *http.Server, l net.Listener, logger *slog.Logger, shutdownTime time.Duration) error {
	errCh := make(chan error, 1)

	go func() {
		err := server.Serve(l)
		if !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "serve", slog.Any("error", err))
		}

		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTime)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "server forced to shutdown", slog.Any("error", err))

			return fmt.Errorf("httpServer.Shutdown: %w", err)
		}

		if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("httpServer.Serve: %w", err)
		}

		return nil
	case err := <-errCh:
		return fmt.Errorf("httpServer.Serve: %w", err)
	}
}
