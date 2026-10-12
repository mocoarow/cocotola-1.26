package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func Test_NewActiveQuestionList_shouldKeepWorkbookID_whenValid(t *testing.T) {
	t.Parallel()

	// given, when
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1"})

	// then
	require.NoError(t, err)
	assert.Equal(t, "wb-1", list.WorkbookID())
}

func Test_NewActiveQuestionList_shouldHoldGivenIDs_whenValid(t *testing.T) {
	t.Parallel()

	// given, when
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-2"})

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{"q-1", "q-2"}, list.Entries())
}

func Test_NewActiveQuestionList_shouldDeduplicateIDs_whenIDsAreDuplicated(t *testing.T) {
	t.Parallel()

	// given, when
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-1", "q-2"})

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, list.Size())
}

func Test_NewActiveQuestionList_shouldStartAtVersionZero_whenCreated(t *testing.T) {
	t.Parallel()

	// given, when
	list, err := domain.NewActiveQuestionList("wb-1", nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, list.Version())
}

func Test_NewActiveQuestionList_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given, when
	_, err := domain.NewActiveQuestionList("", []string{"q-1"})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewActiveQuestionList_shouldReturnError_whenQuestionIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given, when
	_, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", ""})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ActiveQuestionList_Add_shouldContainID_whenIDIsNotEmpty(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1"})
	require.NoError(t, err)

	// when
	err = list.Add("q-2")

	// then
	require.NoError(t, err)
	assert.True(t, list.Contains("q-2"))
}

func Test_ActiveQuestionList_Add_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1"})
	require.NoError(t, err)

	// when
	err = list.Add("")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ActiveQuestionList_Add_shouldKeepSize_whenIDAlreadyExists(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-2"})
	require.NoError(t, err)

	// when
	err = list.Add("q-1")

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, list.Size())
}

func Test_ActiveQuestionList_Size_shouldIncrease_whenNewIDIsAdded(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-2"})
	require.NoError(t, err)

	// when
	err = list.Add("q-3")

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, list.Size())
}

func Test_ActiveQuestionList_Size_shouldDecrease_whenExistingIDIsRemoved(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-2", "q-3"})
	require.NoError(t, err)

	// when
	list.Remove("q-2")

	// then
	assert.Equal(t, 2, list.Size())
}

func Test_ActiveQuestionList_Remove_shouldRemoveID_whenIDExists(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-2"})
	require.NoError(t, err)

	// when
	list.Remove("q-1")

	// then
	assert.Equal(t, []string{"q-2"}, list.Entries())
}

func Test_ActiveQuestionList_Remove_shouldKeepEntries_whenIDNotFound(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1", "q-2"})
	require.NoError(t, err)

	// when
	list.Remove("q-99")

	// then
	assert.Equal(t, []string{"q-1", "q-2"}, list.Entries())
}

func Test_ActiveQuestionList_Entries_shouldReturnSortedIDs_whenIDsAreUnordered(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-c", "q-a", "q-b"})
	require.NoError(t, err)

	// when
	entries := list.Entries()

	// then
	assert.Equal(t, []string{"q-a", "q-b", "q-c"}, entries)
}

func Test_ActiveQuestionList_Entries_shouldNotAffectList_whenReturnedSliceIsModified(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", []string{"q-a", "q-b"})
	require.NoError(t, err)
	entries := list.Entries()

	// when
	entries[0] = "mutated"

	// then
	assert.Equal(t, []string{"q-a", "q-b"}, list.Entries())
}

func Test_ActiveQuestionList_Contains_shouldReportMembership_whenIDIsQueried(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		questionID string
		want       bool
	}{
		{name: "existing id", questionID: "q-1", want: true},
		{name: "missing id", questionID: "q-99", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			list, err := domain.NewActiveQuestionList("wb-1", []string{"q-1"})
			require.NoError(t, err)

			// when
			got := list.Contains(tt.questionID)

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_ActiveQuestionList_SetVersion_shouldUpdateVersion_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	list, err := domain.NewActiveQuestionList("wb-1", nil)
	require.NoError(t, err)

	// when
	list.SetVersion(5)

	// then
	assert.Equal(t, 5, list.Version())
}
