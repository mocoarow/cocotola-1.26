package idset_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"
)

func Test_New_shouldCreateEmptySet_whenNoIDs(t *testing.T) {
	t.Parallel()

	// given / when
	s := idset.New[int, int](1, nil)

	// then
	assert.Equal(t, 1, s.OwnerID)
	assert.Equal(t, 0, s.Size())
}

func Test_New_shouldCreateSetWithEntries_whenIDsProvided(t *testing.T) {
	t.Parallel()

	// given / when
	s := idset.New[int, int](1, []int{10, 20, 30})

	// then
	assert.Equal(t, 3, s.Size())
	assert.True(t, s.Contains(10))
	assert.True(t, s.Contains(20))
	assert.True(t, s.Contains(30))
}

func Test_Set_Contains_shouldReturnFalse_whenIDNotPresent(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10})

	// when / then
	assert.False(t, s.Contains(99))
}

func Test_Set_Add_shouldAddID(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, nil)

	// when
	s.Add(42)

	// then
	assert.True(t, s.Contains(42))
	assert.Equal(t, 1, s.Size())
}

func Test_Set_Remove_shouldRemoveID(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10, 20})

	// when
	s.Remove(10)

	// then
	assert.False(t, s.Contains(10))
	assert.Equal(t, 1, s.Size())
}

var errLimit = errors.New("limit reached")
var errDup = errors.New("duplicate")

func Test_Set_AddWithLimit_shouldSucceed_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10})

	// when
	err := s.AddWithLimit(20, 5, errLimit, errDup)

	// then
	require.NoError(t, err)
	assert.True(t, s.Contains(20))
}

func Test_Set_AddWithLimit_shouldReturnLimitErr_whenAtCapacity(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10, 20, 30})

	// when
	err := s.AddWithLimit(40, 3, errLimit, errDup)

	// then
	require.ErrorIs(t, err, errLimit)
}

func Test_Set_AddWithLimit_shouldReturnDupErr_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10, 20})

	// when
	err := s.AddWithLimit(10, 5, errLimit, errDup)

	// then
	require.ErrorIs(t, err, errDup)
}

func Test_Set_AddUnique_shouldSucceed_whenNotDuplicate(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10})

	// when
	err := s.AddUnique(20, errDup)

	// then
	require.NoError(t, err)
	assert.True(t, s.Contains(20))
}

func Test_Set_AddUnique_shouldReturnDupErr_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	s := idset.New[int, int](1, []int{10})

	// when
	err := s.AddUnique(10, errDup)

	// then
	require.ErrorIs(t, err, errDup)
}
