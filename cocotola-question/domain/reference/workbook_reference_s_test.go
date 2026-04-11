package reference_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
)

func validRefArgs() (string, string, string, time.Time) {
	return "ref-1", "user-1", "wb-1", time.Now()
}

func Test_NewWorkbookReference_shouldReturnReference_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, userID, wbID, addedAt := validRefArgs()

	// when
	ref, err := reference.NewWorkbookReference(id, userID, wbID, addedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, ref.ID())
	assert.Equal(t, userID, ref.UserID())
	assert.Equal(t, wbID, ref.WorkbookID())
	assert.Equal(t, addedAt, ref.AddedAt())
}

func Test_NewWorkbookReference_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, userID, wbID, addedAt := validRefArgs()

	// when
	_, err := reference.NewWorkbookReference("", userID, wbID, addedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbookReference_shouldReturnError_whenUserIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, _, wbID, addedAt := validRefArgs()

	// when
	_, err := reference.NewWorkbookReference(id, "", wbID, addedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbookReference_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, userID, _, addedAt := validRefArgs()

	// when
	_, err := reference.NewWorkbookReference(id, userID, "", addedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ReconstructWorkbookReference_shouldReturnReference_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	id, userID, wbID, addedAt := validRefArgs()

	// when
	ref := reference.ReconstructWorkbookReference(id, userID, wbID, addedAt)

	// then
	assert.Equal(t, id, ref.ID())
	assert.Equal(t, userID, ref.UserID())
	assert.Equal(t, wbID, ref.WorkbookID())
}
