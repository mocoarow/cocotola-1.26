package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func runHTTPServer(ctx context.Context, server *http.Server, logger *slog.Logger, shutdownTime time.Duration) error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "listen and serve", slog.Any("error", err))

			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTime)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "server forced to shutdown", slog.Any("error", err))

			return fmt.Errorf("httpServer.Shutdown: %w", err)
		}

		return nil
	case err := <-errCh:
		return fmt.Errorf("httpServer.ListenAndServe: %w", err)
	}
}
