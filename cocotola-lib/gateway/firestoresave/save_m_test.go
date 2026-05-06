package firestoresave_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
)

const testProjectID = "firestoresave-test"

var docCounter atomic.Uint64

func setupFirestoreClient(t *testing.T) *firestore.Client {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST not set; skipping Firestore integration test")
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, testProjectID)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("close firestore client: %v", err)
		}
	})

	return client
}

func uniqueDocRef(client *firestore.Client) *firestore.DocumentRef {
	id := fmt.Sprintf("doc-%d-%d", time.Now().UnixNano(), docCounter.Add(1))
	return client.Collection("firestoresave_test").Doc(id)
}

func decodeFakeRecordVersion(snap *firestore.DocumentSnapshot) (int, error) {
	var rec fakeRecord
	if err := snap.DataTo(&rec); err != nil {
		return 0, fmt.Errorf("decode fake record: %w", err)
	}
	return rec.Version, nil
}

func Test_SaveVersioned_shouldInsertDocument_whenEntityVersionIsZeroAndDocAbsent(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)
	entity := &fakeEntity{version: 0}

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     entity,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "first", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	})

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, entity.Version())

	snap, err := docRef.Get(ctx)
	require.NoError(t, err)
	var got fakeRecord
	require.NoError(t, snap.DataTo(&got))
	assert.Equal(t, "first", got.ID)
	assert.Equal(t, 1, got.Version)
}

func Test_SaveVersioned_shouldUpdateDocument_whenVersionMatches(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)

	// seed
	require.NoError(t, firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     &fakeEntity{version: 0},
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "v1", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	}))

	loaded := &fakeEntity{version: 1}

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     loaded,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "v2", Version: 2},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	})

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, loaded.Version())

	snap, err := docRef.Get(ctx)
	require.NoError(t, err)
	var got fakeRecord
	require.NoError(t, snap.DataTo(&got))
	assert.Equal(t, "v2", got.ID)
	assert.Equal(t, 2, got.Version)
}

func Test_SaveVersioned_shouldReturnConcurrentModification_whenVersionMismatch(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)

	// seed at version 1
	require.NoError(t, firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     &fakeEntity{version: 0},
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "v1", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	}))

	// advance to version 2 via the "winner"
	require.NoError(t, firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     &fakeEntity{version: 1},
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "winner", Version: 2},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	}))

	// stale loser still thinks the document is at version 1
	loser := &fakeEntity{version: 1}

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     loser,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "loser", Version: 2},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	})

	// then
	require.ErrorIs(t, err, versioned.ErrConcurrentModification)
	assert.Equal(t, 1, loser.Version(), "entity version must not be advanced on failure")

	snap, err := docRef.Get(ctx)
	require.NoError(t, err)
	var got fakeRecord
	require.NoError(t, snap.DataTo(&got))
	assert.Equal(t, "winner", got.ID, "winner's write must not be overwritten by loser")
}

func Test_SaveVersioned_shouldReturnConcurrentModification_whenInsertAttemptedAndDocExists(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)

	// seed at version 1 via a fresh entity
	require.NoError(t, firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     &fakeEntity{version: 0},
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "first", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	}))

	// another caller also believes this is a fresh aggregate
	freshDuplicate := &fakeEntity{version: 0}

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     freshDuplicate,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "dup", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	})

	// then
	require.ErrorIs(t, err, versioned.ErrConcurrentModification)
	assert.Equal(t, 0, freshDuplicate.Version(), "entity version must not be advanced on failure")

	snap, err := docRef.Get(ctx)
	require.NoError(t, err)
	var got fakeRecord
	require.NoError(t, snap.DataTo(&got))
	assert.Equal(t, "first", got.ID, "existing doc must not be overwritten by a duplicate insert")
}

func Test_SaveVersioned_shouldReturnConcurrentModification_whenInsertAttemptedAndDocExistsWithStoredVersionZero(t *testing.T) {
	t.Parallel()
	// given: a doc exists at the target ref but its stored version is 0
	// (a corrupt/legacy record), and a fresh aggregate (entity.Version()==0)
	// attempts to insert. The helper must refuse to overwrite, regardless of
	// the stored version, since both sides claim "fresh".
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)

	_, err := docRef.Set(ctx, &fakeRecord{ID: "preexisting-zero", Version: 0})
	require.NoError(t, err)

	fresh := &fakeEntity{version: 0}

	// when
	err = firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     fresh,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "would-overwrite", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	})

	// then
	require.ErrorIs(t, err, versioned.ErrConcurrentModification)
	assert.Equal(t, 0, fresh.Version(), "entity version must not be advanced on failure")

	snap, err := docRef.Get(ctx)
	require.NoError(t, err)
	var got fakeRecord
	require.NoError(t, snap.DataTo(&got))
	assert.Equal(t, "preexisting-zero", got.ID, "existing doc must not be overwritten")
}

func Test_SaveVersioned_shouldReturnNotFound_whenDocAbsentAndEntityVersionIsNonZero(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)
	entity := &fakeEntity{version: 5}

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     entity,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "ghost", Version: 6},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	})

	// then
	require.ErrorIs(t, err, versioned.ErrNotFound)
	assert.NotErrorIs(t, err, versioned.ErrConcurrentModification,
		"NotFound and ConcurrentModification must remain distinct")
}

func Test_SaveVersioned_shouldWrapDecodeError_whenDecodeReturnsError(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	client := setupFirestoreClient(t)
	docRef := uniqueDocRef(client)

	// seed something at version 1
	require.NoError(t, firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     &fakeEntity{version: 0},
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "seeded", Version: 1},
		Decode:     decodeFakeRecordVersion,
		EntityName: "fake",
	}))

	loaded := &fakeEntity{version: 1}
	failingDecode := func(*firestore.DocumentSnapshot) (int, error) {
		return 0, errors.New("boom")
	}

	// when
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*fakeRecord]{
		Client:     client,
		Entity:     loaded,
		DocRef:     docRef,
		NewRecord:  &fakeRecord{ID: "v2", Version: 2},
		Decode:     failingDecode,
		EntityName: "fake",
	})

	// then
	require.ErrorContains(t, err, "decode fake")
	require.ErrorContains(t, err, "boom")
}
