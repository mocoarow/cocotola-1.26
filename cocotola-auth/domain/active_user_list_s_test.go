package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_NewActiveUserList_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewActiveUserList(0, nil)

	// then
	assert.Error(t, err)
}

func Test_NewActiveUserList_shouldReturnError_whenOrganizationIDIsNegative(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewActiveUserList(-1, nil)

	// then
	assert.Error(t, err)
}

func Test_ActiveUserList_Add_shouldSucceed_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, []int{1, 2})

	// when
	err := list.Add(3, 5)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(3))
}

func Test_ActiveUserList_Add_shouldReturnError_whenAtLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, []int{1, 2, 3})

	// when
	err := list.Add(4, 3)

	// then
	assert.ErrorIs(t, err, domain.ErrActiveUserLimitReached)
}

func Test_ActiveUserList_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, []int{1, 2})

	// when
	err := list.Add(2, 5)

	// then
	assert.ErrorIs(t, err, domain.ErrDuplicateEntry)
}

func Test_ActiveUserList_Remove_shouldRemoveEntry(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, []int{1, 2, 3})

	// when
	list.Remove(2)

	// then
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(2))
}

func Test_ActiveUserList_Remove_shouldDoNothing_whenIDNotFound(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, []int{1, 2})

	// when
	list.Remove(99)

	// then
	assert.Equal(t, 2, list.Size())
}

func Test_ActiveUserList_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, nil)

	// when
	result := list.Contains(1)

	// then
	assert.False(t, result)
}

func Test_ActiveUserList_Add_shouldSucceed_whenAddingToEmptyList(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewActiveUserList(1, nil)

	// when
	err := list.Add(1, 5)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1, list.Size())
}
