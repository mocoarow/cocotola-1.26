package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var (
	fixtureUser1 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000021")
	fixtureUser2 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000022")
	fixtureUser3 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000023")
	fixtureUser4 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000024")
	fixtureUser5 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000025")
)

func Test_NewActiveUserList_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewActiveUserList(domain.OrganizationID{}, nil)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ActiveUserList_Add_shouldSucceed_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, []domain.AppUserID{fixtureUser1, fixtureUser2})

	// when
	err := list.Add(fixtureUser3, 5)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(fixtureUser3))
}

func Test_ActiveUserList_Add_shouldReturnError_whenAtLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, []domain.AppUserID{fixtureUser1, fixtureUser2, fixtureUser3})

	// when
	err := list.Add(fixtureUser4, 3)

	// then
	require.ErrorIs(t, err, domain.ErrActiveUserLimitReached)
}

func Test_ActiveUserList_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, []domain.AppUserID{fixtureUser1, fixtureUser2})

	// when
	err := list.Add(fixtureUser2, 5)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_ActiveUserList_Remove_shouldRemoveEntry(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, []domain.AppUserID{fixtureUser1, fixtureUser2, fixtureUser3})

	// when
	list.Remove(fixtureUser2)

	// then
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(fixtureUser2))
}

func Test_ActiveUserList_Remove_shouldDoNothing_whenIDNotFound(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, []domain.AppUserID{fixtureUser1, fixtureUser2})

	// when
	list.Remove(fixtureUser5)

	// then
	assert.Equal(t, 2, list.Size())
}

func Test_ActiveUserList_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, nil)

	// when
	result := list.Contains(fixtureUser1)

	// then
	assert.False(t, result)
}

func Test_ActiveUserList_Add_shouldSucceed_whenAddingToEmptyList(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(fixtureOrgID, nil)

	// when
	err := list.Add(fixtureUser1, 5)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, list.Size())
}
