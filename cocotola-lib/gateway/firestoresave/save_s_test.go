package firestoresave_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
)

func Test_SaveVersioned_shouldReturnError_whenRecordVersionDoesNotMatchNextVersion(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     nil,
		Entity:     &fakeEntity{version: 3},
		DocRef:     nil,
		NewRecord:  &fakeRecord{ID: "x", Version: 3},
		Decode:     nil,
		EntityName: "question",
	})

	// then
	require.ErrorContains(t, err, "record Version=3 does not match expected next version 4")
}

func Test_SaveVersioned_shouldReturnError_whenRecordVersionIsZeroOnInsert(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     nil,
		Entity:     &fakeEntity{version: 0},
		DocRef:     nil,
		NewRecord:  &fakeRecord{ID: "x", Version: 0},
		Decode:     nil,
		EntityName: "question",
	})

	// then
	require.ErrorContains(t, err, "record Version=0 does not match expected next version 1")
}

func Test_SaveVersioned_shouldEmbedEntityNameInError_whenValidationFails(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     nil,
		Entity:     &fakeEntity{version: 0},
		DocRef:     nil,
		NewRecord:  &fakeRecord{ID: "x", Version: 0},
		Decode:     nil,
		EntityName: "active question list",
	})

	// then
	require.ErrorContains(t, err, "save active question list")
}
