package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

var (
	fixtureGroupID1  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000001")
	fixtureGroupID2  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000002")
	fixtureGroupID3  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000003")
	fixtureGroupID4  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000004")
	fixtureGroupID99 = domain.MustParseGroupID("00000000-0000-7000-8000-000000000099")
)

func Test_NewGroupChildGroups_shouldReturnError_whenGroupIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := group.NewChildGroups(domain.GroupID{}, nil)

	// then
	require.Error(t, err)
}

func Test_GroupChildGroups_Add_shouldSucceed_whenGroupNotInChildren(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewChildGroups(fixtureGroupID1, []domain.GroupID{fixtureGroupID2, fixtureGroupID3})

	// when
	err := g.Add(fixtureGroupID4)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, g.Size())
	assert.True(t, g.Contains(fixtureGroupID4))
}

func Test_GroupChildGroups_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewChildGroups(fixtureGroupID1, []domain.GroupID{fixtureGroupID2, fixtureGroupID3})

	// when
	err := g.Add(fixtureGroupID3)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_GroupChildGroups_Remove_shouldRemoveChildGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewChildGroups(fixtureGroupID1, []domain.GroupID{fixtureGroupID2, fixtureGroupID3, fixtureGroupID4})

	// when
	g.Remove(fixtureGroupID3)

	// then
	assert.Equal(t, 2, g.Size())
	assert.False(t, g.Contains(fixtureGroupID3))
}

func Test_GroupChildGroups_Remove_shouldDoNothing_whenGroupNotInChildren(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewChildGroups(fixtureGroupID1, []domain.GroupID{fixtureGroupID2, fixtureGroupID3})

	// when
	g.Remove(fixtureGroupID99)

	// then
	assert.Equal(t, 2, g.Size())
}

func Test_GroupChildGroups_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewChildGroups(fixtureGroupID1, nil)

	// when
	result := g.Contains(fixtureGroupID2)

	// then
	assert.False(t, result)
}

func Test_GroupChildGroups_Add_shouldSucceed_whenAddingToEmptyGroup(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewChildGroups(fixtureGroupID1, nil)

	// when
	err := g.Add(fixtureGroupID2)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, g.Size())
}
