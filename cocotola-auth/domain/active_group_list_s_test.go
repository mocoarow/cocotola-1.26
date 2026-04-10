package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var fixtureOrgID = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")

func Test_NewActiveGroupList_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewActiveGroupList(domain.OrganizationID{}, nil)

	// then
	require.Error(t, err)
}

func Test_ActiveGroupList_Add_shouldSucceed_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []int{1, 2})

	// when
	err := list.Add(3, 5)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(3))
}

func Test_ActiveGroupList_Add_shouldReturnError_whenAtLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []int{1, 2, 3})

	// when
	err := list.Add(4, 3)

	// then
	require.ErrorIs(t, err, domain.ErrActiveGroupLimitReached)
}

func Test_ActiveGroupList_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []int{1, 2})

	// when
	err := list.Add(2, 5)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_ActiveGroupList_Remove_shouldRemoveEntry(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []int{1, 2, 3})

	// when
	list.Remove(2)

	// then
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(2))
}

func Test_ActiveGroupList_Remove_shouldDoNothing_whenIDNotFound(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []int{1, 2})

	// when
	list.Remove(99)

	// then
	assert.Equal(t, 2, list.Size())
}

func Test_ActiveGroupList_Add_shouldSucceed_whenAddingToEmptyList(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, nil)

	// when
	err := list.Add(1, 5)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, list.Size())
}
