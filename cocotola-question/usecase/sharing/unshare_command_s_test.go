package sharing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
)

const fixtureReferenceID = "wb-ref-1"

func newUnshareInput(t *testing.T) *referenceservice.UnshareInput {
	t.Helper()
	input, err := referenceservice.NewUnshareInput(fixtureOperatorID, fixtureOrganizationID, fixtureReferenceID)
	require.NoError(t, err)
	return input
}

func Test_UnshareCommand_shouldDeleteOperatorReference_whenDeleterSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	deleter := newMockreferenceDeleter(t)
	deleter.On("Delete", mock.Anything, fixtureOperatorID, fixtureReferenceID).Return(nil)
	cmd := sharingusecase.NewUnshareCommand(deleter)

	// when
	err := cmd.Unshare(ctx, newUnshareInput(t))

	// then
	require.NoError(t, err)
}

func Test_UnshareCommand_shouldReturnError_whenDeleterFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	deleterErr := errors.New("firestore unavailable")
	deleter := newMockreferenceDeleter(t)
	deleter.On("Delete", mock.Anything, fixtureOperatorID, fixtureReferenceID).Return(deleterErr)
	cmd := sharingusecase.NewUnshareCommand(deleter)

	// when
	err := cmd.Unshare(ctx, newUnshareInput(t))

	// then
	require.ErrorIs(t, err, deleterErr)
}
