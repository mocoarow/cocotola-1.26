package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func Test_NewOwnedWorkbookList_shouldReturnList_whenValid(t *testing.T) {
	t.Parallel()

	// given

	// when
	list, err := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", "wb-2"})

	// then
	require.NoError(t, err)
	assert.Equal(t, "user-1", list.OwnerID())
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains("wb-1"))
	assert.True(t, list.Contains("wb-2"))
}

func Test_NewOwnedWorkbookList_shouldReturnError_whenOwnerIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given

	// when
	_, err := domain.NewOwnedWorkbookList("", nil)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewOwnedWorkbookList_shouldDeduplicateIDs(t *testing.T) {
	t.Parallel()

	// given

	// when
	list, err := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", "wb-1", "wb-2"})

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, list.Size())
}

func Test_NewOwnedWorkbookList_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given

	// when
	_, err := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", ""})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_OwnedWorkbookList_Add_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-1"})

	// when
	err := list.Add("", 3)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_OwnedWorkbookList_Add_shouldSucceed_whenUnderLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-1"})

	// when
	err := list.Add("wb-2", 3)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains("wb-2"))
}

func Test_OwnedWorkbookList_Add_shouldReturnError_whenAtLimit(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", "wb-2", "wb-3"})

	// when
	err := list.Add("wb-4", 3)

	// then
	require.ErrorIs(t, err, domain.ErrOwnedWorkbookLimitReached)
}

func Test_OwnedWorkbookList_Add_shouldReturnError_whenDuplicate(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", "wb-2"})

	// when
	err := list.Add("wb-1", 5)

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateOwnedWorkbook)
}

func Test_OwnedWorkbookList_Remove_shouldRemoveEntry(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", "wb-2", "wb-3"})

	// when
	list.Remove("wb-2")

	// then
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains("wb-2"))
}

func Test_OwnedWorkbookList_Remove_shouldDoNothing_whenIDNotFound(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-1", "wb-2"})

	// when
	list.Remove("wb-99")

	// then
	assert.Equal(t, 2, list.Size())
}

func Test_OwnedWorkbookList_Contains_shouldReturnFalse_whenEmpty(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", nil)

	// when
	result := list.Contains("wb-1")

	// then
	assert.False(t, result)
}

func Test_OwnedWorkbookList_Add_shouldSucceed_whenAddingToEmptyList(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", nil)

	// when
	err := list.Add("wb-1", 3)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, list.Size())
}

func Test_OwnedWorkbookList_Entries_shouldReturnSortedCopy(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", []string{"wb-c", "wb-a", "wb-b"})

	// when
	entries := list.Entries()

	// then
	assert.Equal(t, []string{"wb-a", "wb-b", "wb-c"}, entries)

	// Mutating the returned slice must not affect the aggregate.
	entries[0] = "mutated"
	assert.True(t, list.Contains("wb-a"))
}

func Test_OwnedWorkbookList_Version_shouldDefaultToZero(t *testing.T) {
	t.Parallel()

	// given

	// when
	list, _ := domain.NewOwnedWorkbookList("user-1", nil)

	// then
	assert.Equal(t, 0, list.Version())
}

func Test_OwnedWorkbookList_SetVersion_shouldSetVersion(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", nil)

	// when
	list.SetVersion(5)

	// then
	assert.Equal(t, 5, list.Version())
}

func Test_OwnedWorkbookList_Add_shouldReturnError_whenMaxWorkbooksIsZero(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", nil)

	// when
	err := list.Add("wb-1", 0)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_OwnedWorkbookList_Add_shouldReturnError_whenMaxWorkbooksIsNegative(t *testing.T) {
	t.Parallel()

	// given
	list, _ := domain.NewOwnedWorkbookList("user-1", nil)

	// when
	err := list.Add("wb-1", -1)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
