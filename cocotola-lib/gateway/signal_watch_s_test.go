package gateway_test

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

// A signal sent to the process reaches every subscribed channel, so tests that send or subscribe to real signals hold this lock.
var processSignalMu sync.Mutex

func Test_SignalWatchProcess_shouldReturnContextCanceledError_whenContextCanceled(t *testing.T) {
	t.Parallel()

	processSignalMu.Lock()
	defer processSignalMu.Unlock()

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

const signalDeliveryTimeout = 5 * time.Second

func Test_SignalWatchProcess_shouldKeepOtherSignalSubscriptions_whenReturned(t *testing.T) {
	t.Parallel()

	processSignalMu.Lock()
	defer processSignalMu.Unlock()

	// given
	other := make(chan os.Signal, 1)
	signal.Notify(other, syscall.SIGINT)
	defer signal.Stop(other)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, gateway.SignalWatchProcess(ctx), context.Canceled)

	// when
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))

	// then
	select {
	case <-other:
	case <-time.After(signalDeliveryTimeout):
		t.Fatal("SIGINT was not delivered to the other subscription")
	}
}
