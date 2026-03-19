package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_NewGroupChildGroups_shouldReturnError_whenGroupIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewGroupChildGroups(0, nil)

	// then
	require.Error(t, err)
}

func Test_NewGroupChildGroups_shouldReturnError_whenGroupIDIsNegative(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewGroupChildGroups(-1, nil)

	// then
	require.Error(t, err)
}

func Test_GroupChildGroups_Add_shouldSucceed_whenGroupNotInChildren(t *testing.T) {
	t.Parallel()

	// given
	g, _ := domain.NewGroupChildGroups(1, []int{2, 3})

	// when
	err := g.Add(4)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, g.Size())
	assert.True(t, g.Contains(4))
}

func Test_GroupChildGroups_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	g, _ := domain.NewGroupChildGroups(1, []int{2, 3})

	// when
	err := g.Add(3)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_GroupChildGroups_Remove_shouldRemoveChildGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := domain.NewGroupChildGroups(1, []int{2, 3, 4})

	// when
	g.Remove(3)

	// then
	assert.Equal(t, 2, g.Size())
	assert.False(t, g.Contains(3))
}

func Test_GroupChildGroups_Remove_shouldDoNothing_whenGroupNotInChildren(t *testing.T) {
	t.Parallel()

	// given
	g, _ := domain.NewGroupChildGroups(1, []int{2, 3})

	// when
	g.Remove(99)

	// then
	assert.Equal(t, 2, g.Size())
}

func Test_GroupChildGroups_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	g, _ := domain.NewGroupChildGroups(1, nil)

	// when
	result := g.Contains(2)

	// then
	assert.False(t, result)
}

func Test_GroupChildGroups_Add_shouldSucceed_whenAddingToEmptyGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := domain.NewGroupChildGroups(1, nil)

	// when
	err := g.Add(2)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, g.Size())
}
