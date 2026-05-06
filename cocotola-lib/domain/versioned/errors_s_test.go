package versioned_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

func Test_ErrConcurrentModification_shouldHaveStableMessage(t *testing.T) {
	t.Parallel()
	// given
	err := versioned.ErrConcurrentModification

	// when
	msg := err.Error()

	// then
	assert.Equal(t, "concurrent modification", msg)
}

func Test_ErrConcurrentModification_shouldBeIdentifiableViaErrorsIs_whenWrapped(t *testing.T) {
	t.Parallel()
	// given
	wrapped := fmt.Errorf("save user: %w", versioned.ErrConcurrentModification)

	// when
	matched := errors.Is(wrapped, versioned.ErrConcurrentModification)

	// then
	require.True(t, matched)
}

func Test_ErrNotFound_shouldHaveStableMessage(t *testing.T) {
	t.Parallel()
	// given
	err := versioned.ErrNotFound

	// when
	msg := err.Error()

	// then
	assert.Equal(t, "versioned entity not found", msg)
}

func Test_ErrNotFound_shouldBeIdentifiableViaErrorsIs_whenWrapped(t *testing.T) {
	t.Parallel()
	// given
	wrapped := fmt.Errorf("save user: %w", versioned.ErrNotFound)

	// when
	matched := errors.Is(wrapped, versioned.ErrNotFound)

	// then
	require.True(t, matched)
}

func Test_ErrNotFound_shouldNotMatchErrConcurrentModification(t *testing.T) {
	t.Parallel()
	// given
	err := versioned.ErrNotFound

	// when
	matched := errors.Is(err, versioned.ErrConcurrentModification)

	// then
	assert.False(t, matched)
}

type fakeEntity struct {
	version int
}

func (e *fakeEntity) Version() int     { return e.version }
func (e *fakeEntity) SetVersion(v int) { e.version = v }

func Test_Entity_shouldBeImplementableByConcreteType(t *testing.T) {
	t.Parallel()
	// given
	var entity versioned.Entity = &fakeEntity{version: 3}

	// when
	entity.SetVersion(entity.Version() + 1)

	// then
	assert.Equal(t, 4, entity.Version())
}

type fakeRecord struct {
	version int
}

func (r *fakeRecord) GetVersion() int { return r.version }

func Test_Record_shouldBeImplementableByConcreteType(t *testing.T) {
	t.Parallel()
	// given
	var record versioned.Record = &fakeRecord{version: 7}

	// when
	got := record.GetVersion()

	// then
	assert.Equal(t, 7, got)
}
