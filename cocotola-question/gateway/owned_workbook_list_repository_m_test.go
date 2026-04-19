package gateway_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const testProjectID = "test-project"

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

func Test_OwnedWorkbookListRepository_FindByOwnerID_shouldReturnEmptyList_whenOwnerNotExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewOwnedWorkbookListRepository(client)

	// when
	list, err := repo.FindByOwnerID(ctx, "nonexistent-owner")

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, list.Size())
	assert.Equal(t, 0, list.Version())
}

func Test_OwnedWorkbookListRepository_SaveAndFind_shouldPersistEntries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewOwnedWorkbookListRepository(client)
	ownerID := "test-owner-save-" + t.Name()

	list, err := domain.NewOwnedWorkbookList(ownerID, []string{"wb-1", "wb-2"})
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, list)

	// then
	require.NoError(t, err)

	// when - reload
	loaded, err := repo.FindByOwnerID(ctx, ownerID)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, loaded.Size())
	assert.True(t, loaded.Contains("wb-1"))
	assert.True(t, loaded.Contains("wb-2"))
	assert.Equal(t, 1, loaded.Version())
}

func Test_OwnedWorkbookListRepository_Save_shouldReturnConcurrentModification_whenVersionMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewOwnedWorkbookListRepository(client)
	ownerID := "test-owner-concurrent-" + t.Name()

	list1, _ := domain.NewOwnedWorkbookList(ownerID, []string{"wb-1"})
	require.NoError(t, repo.Save(ctx, list1))

	// Simulate stale version: create a list with version 0 (stale).
	staleList, _ := domain.NewOwnedWorkbookList(ownerID, []string{"wb-1", "wb-2"})
	// staleList has version 0, but DB has version 1.

	// when
	err := repo.Save(ctx, staleList)

	// then
	require.ErrorIs(t, err, domain.ErrConcurrentModification)
}
