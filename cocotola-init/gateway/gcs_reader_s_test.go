//go:build small

package gateway_test

import (
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/gateway"
)

func Test_ReadAllWithCap_shouldReturnAllBytes_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	r := strings.NewReader("hello")

	// when
	data, err := gateway.ReadAllWithCap(r, 10)

	// then
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func Test_ReadAllWithCap_shouldReturnAllBytes_whenExactlyAtLimit(t *testing.T) {
	t.Parallel()

	// given: the source length equals the cap exactly
	r := strings.NewReader("0123456789")

	// when
	data, err := gateway.ReadAllWithCap(r, 10)

	// then
	require.NoError(t, err)
	assert.Equal(t, "0123456789", string(data))
}

func Test_ReadAllWithCap_shouldReturnError_whenOverLimit(t *testing.T) {
	t.Parallel()

	// given: one byte more than the cap
	r := strings.NewReader("0123456789X")

	// when
	_, err := gateway.ReadAllWithCap(r, 10)

	// then
	require.ErrorIs(t, err, gateway.ErrObjectTooLarge)
}

func Test_ReadAllWithCap_shouldReturnError_whenReaderFails(t *testing.T) {
	t.Parallel()

	// given
	readErr := iotest.ErrReader(assert.AnError)

	// when
	_, err := gateway.ReadAllWithCap(readErr, 10)

	// then
	require.ErrorIs(t, err, assert.AnError)
}
