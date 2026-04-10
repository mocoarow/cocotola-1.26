package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

var (
	fixtureGroupUser1 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000031")
	fixtureGroupUser2 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000032")
	fixtureGroupUser3 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000033")
	fixtureGroupUser4 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000034")
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
	g, _ := group.NewUsers(1, []domain.AppUserID{fixtureGroupUser1, fixtureGroupUser2})

	// when
	err := g.Add(fixtureGroupUser3)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, g.Size())
	assert.True(t, g.Contains(fixtureGroupUser3))
}

func Test_GroupUsers_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []domain.AppUserID{fixtureGroupUser1, fixtureGroupUser2})

	// when
	err := g.Add(fixtureGroupUser2)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_GroupUsers_Remove_shouldRemoveUser(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []domain.AppUserID{fixtureGroupUser1, fixtureGroupUser2, fixtureGroupUser3})

	// when
	g.Remove(fixtureGroupUser2)

	// then
	assert.Equal(t, 2, g.Size())
	assert.False(t, g.Contains(fixtureGroupUser2))
}

func Test_GroupUsers_Remove_shouldDoNothing_whenUserNotInGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, []domain.AppUserID{fixtureGroupUser1, fixtureGroupUser2})

	// when
	g.Remove(fixtureGroupUser4)

	// then
	assert.Equal(t, 2, g.Size())
}

func Test_GroupUsers_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, nil)

	// when
	result := g.Contains(fixtureGroupUser1)

	// then
	assert.False(t, result)
}

func Test_GroupUsers_Add_shouldSucceed_whenAddingToEmptyGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewUsers(1, nil)

	// when
	err := g.Add(fixtureGroupUser1)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, g.Size())
}
