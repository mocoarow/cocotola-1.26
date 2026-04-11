package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var (
	fixtureOrgID = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")

	fixtureActiveGroupID1  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000001")
	fixtureActiveGroupID2  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000002")
	fixtureActiveGroupID3  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000003")
	fixtureActiveGroupID4  = domain.MustParseGroupID("00000000-0000-7000-8000-000000000004")
	fixtureActiveGroupID99 = domain.MustParseGroupID("00000000-0000-7000-8000-000000000099")
)

func Test_NewActiveGroupList_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewActiveGroupList(domain.OrganizationID{}, nil)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ActiveGroupList_Add_shouldSucceed_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []domain.GroupID{fixtureActiveGroupID1, fixtureActiveGroupID2})

	// when
	err := list.Add(fixtureActiveGroupID3, 5)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(fixtureActiveGroupID3))
}

func Test_ActiveGroupList_Add_shouldReturnError_whenAtLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []domain.GroupID{fixtureActiveGroupID1, fixtureActiveGroupID2, fixtureActiveGroupID3})

	// when
	err := list.Add(fixtureActiveGroupID4, 3)

	// then
	require.ErrorIs(t, err, domain.ErrActiveGroupLimitReached)
}

func Test_ActiveGroupList_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []domain.GroupID{fixtureActiveGroupID1, fixtureActiveGroupID2})

	// when
	err := list.Add(fixtureActiveGroupID2, 5)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_ActiveGroupList_Remove_shouldRemoveEntry(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []domain.GroupID{fixtureActiveGroupID1, fixtureActiveGroupID2, fixtureActiveGroupID3})

	// when
	list.Remove(fixtureActiveGroupID2)

	// then
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(fixtureActiveGroupID2))
}

func Test_ActiveGroupList_Remove_shouldDoNothing_whenIDNotFound(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, []domain.GroupID{fixtureActiveGroupID1, fixtureActiveGroupID2})

	// when
	list.Remove(fixtureActiveGroupID99)

	// then
	assert.Equal(t, 2, list.Size())
}

func Test_ActiveGroupList_Add_shouldSucceed_whenAddingToEmptyList(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveGroupList(fixtureOrgID, nil)

	// when
	err := list.Add(fixtureActiveGroupID1, 5)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, list.Size())
}
