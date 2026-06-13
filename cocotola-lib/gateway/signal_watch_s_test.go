package gateway_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

func Test_SignalWatchProcess_shouldReturnContextCanceledError_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when
	err := gateway.SignalWatchProcess(ctx)

	// then
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled), "expected error to wrap context.Canceled")
}
