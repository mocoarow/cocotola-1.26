package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

func Test_NewGroupUsers_shouldReturnError_whenGroupIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewUsers(0, nil)

	// then
	require.Error(t, err)
}

func Test_NewGroupUsers_shouldReturnError_whenGroupIDIsNegative(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewUsers(-1, nil)

	// then
	require.Error(t, err)
}

func Test_GroupUsers_Add_shouldSucceed_whenUserNotInGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []int{1, 2})

	// when
	err := g.Add(3)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, g.Size())
	assert.True(t, g.Contains(3))
}

func Test_GroupUsers_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []int{1, 2})

	// when
	err := g.Add(2)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_GroupUsers_Remove_shouldRemoveUser(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []int{1, 2, 3})

	// when
	g.Remove(2)

	// then
	assert.Equal(t, 2, g.Size())
	assert.False(t, g.Contains(2))
}

func Test_GroupUsers_Remove_shouldDoNothing_whenUserNotInGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []int{1, 2})

	// when
	g.Remove(99)

	// then
	assert.Equal(t, 2, g.Size())
}

func Test_GroupUsers_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, nil)

	// when
	result := g.Contains(1)

	// then
	assert.False(t, result)
}

func Test_GroupUsers_Add_shouldSucceed_whenAddingToEmptyGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, nil)

	// when
	err := g.Add(1)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, g.Size())
}
