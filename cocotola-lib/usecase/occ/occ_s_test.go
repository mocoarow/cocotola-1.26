package occ_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"
)

var errOther = errors.New("other")

type countingFn struct {
	calls   int
	results []error
}

func (c *countingFn) run() error {
	c.calls++
	if c.calls <= len(c.results) {
		return c.results[c.calls-1]
	}
	return c.results[len(c.results)-1]
}

func Test_Retry_shouldReturnNil_whenFirstAttemptSucceeds(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{nil}}

	// when
	err := occ.Retry(fn.run)

	// then
	require.NoError(t, err)
}

func Test_Retry_shouldReturnNil_whenSecondAttemptSucceedsAfterConflict(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{versioned.ErrConcurrentModification, nil}}

	// when
	err := occ.Retry(fn.run)

	// then
	require.NoError(t, err)
}

func Test_Retry_shouldRunAgain_whenAttemptReturnsWrappedConflict(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{fmt.Errorf("save: %w", versioned.ErrConcurrentModification), nil}}

	// when
	_ = occ.Retry(fn.run)

	// then
	assert.Equal(t, 2, fn.calls)
}

func Test_Retry_shouldReturnConflict_whenEveryAttemptConflicts(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{versioned.ErrConcurrentModification}}

	// when
	err := occ.Retry(fn.run)

	// then
	require.ErrorIs(t, err, versioned.ErrConcurrentModification)
}

func Test_Retry_shouldStopAfterMaxAttempts_whenEveryAttemptConflicts(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{versioned.ErrConcurrentModification}}

	// when
	_ = occ.Retry(fn.run)

	// then
	assert.Equal(t, occ.MaxAttempts, fn.calls)
}

func Test_Retry_shouldReturnError_whenAttemptFailsWithOtherError(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{errOther}}

	// when
	err := occ.Retry(fn.run)

	// then
	require.ErrorIs(t, err, errOther)
}

func Test_Retry_shouldNotRunAgain_whenAttemptFailsWithOtherError(t *testing.T) {
	t.Parallel()

	// given
	fn := &countingFn{results: []error{errOther}}

	// when
	_ = occ.Retry(fn.run)

	// then
	assert.Equal(t, 1, fn.calls)
}
