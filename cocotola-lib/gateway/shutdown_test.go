package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

func Test_newShutdownCtx_shouldReturnContextWithTimeout_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	timeout := 100 * time.Millisecond

	// when
	shutdownCtx, cancel := gateway.NewShutdownCtxForTest(ctx, timeout)
	defer cancel()

	// then
	deadline, ok := shutdownCtx.Deadline()
	require.True(t, ok, "context should have a deadline")
	assert.WithinDuration(t, time.Now().Add(timeout), deadline, 50*time.Millisecond)
}

func Test_newShutdownCtx_shouldStripCancellation_whenParentIsCancelled(t *testing.T) {
	t.Parallel()

	// given
	parent, parentCancel := context.WithCancel(context.Background())

	// when
	shutdownCtx, cancel := gateway.NewShutdownCtxForTest(parent, time.Second)
	defer cancel()
	parentCancel()

	// then
	assert.NoError(t, shutdownCtx.Err(), "shutdown context should not be cancelled when parent is cancelled")
}
