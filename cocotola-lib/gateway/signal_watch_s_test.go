package gateway_test

import (
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

func Test_SignalWatchProcess_shouldReturnContextCanceledError_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when
	err := gateway.SignalWatchProcess(ctx)

	// then
	require.ErrorIs(t, err, context.Canceled)
}

func Test_watchSignal_shouldReturnErrGracefulShutdown_whenSignalReceived(t *testing.T) {
	t.Parallel()

	// given
	sigs := make(chan os.Signal, 1)
	sigs <- syscall.SIGTERM

	// when
	err := gateway.WatchSignal(context.Background(), sigs)

	// then
	require.ErrorIs(t, err, process.ErrGracefulShutdown)
}

func Test_watchSignal_shouldNotReturnContextCanceled_whenSignalReceived(t *testing.T) {
	t.Parallel()

	// given
	sigs := make(chan os.Signal, 1)
	sigs <- syscall.SIGINT

	// when
	err := gateway.WatchSignal(context.Background(), sigs)

	// then
	require.NotErrorIs(t, err, context.Canceled)
}
