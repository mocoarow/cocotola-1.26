package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

// WithSignalWatchProcess returns a RunProcessFunc that watches for OS termination signals.
func WithSignalWatchProcess() process.RunProcessFunc {
	return func(ctx context.Context) process.RunProcess {
		return func() error {
			return SignalWatchProcess(ctx)
		}
	}
}

// SignalWatchProcess blocks until a SIGINT/SIGTERM signal is received or the context is canceled.
func SignalWatchProcess(ctx context.Context) error {
	logger := slog.Default().With(slog.String(logging.LoggerNameKey, "SignalWatch"))
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled: %w", ctx.Err())
	case sig := <-sigs:
		logger.InfoContext(ctx, "shutdown signal received", slog.String("signal", sig.String()))
		return fmt.Errorf("signal %s received: %w", sig.String(), context.Canceled)
	}
}
