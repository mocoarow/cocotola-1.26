package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const activeQuestionListsCollectionForTest = "active_question_lists"

func Test_ActiveQuestionListRepository_FindByWorkbookID_shouldReturnEmptyList_whenWorkbookNotExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewActiveQuestionListRepository(client)

	// when
	list, err := repo.FindByWorkbookID(ctx, "nonexistent-workbook-"+t.Name())

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, list.Size())
	assert.Equal(t, 0, list.Version())
}

func Test_ActiveQuestionListRepository_SaveAndFind_shouldPersistEntries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-save-" + t.Name()

	list, err := domain.NewActiveQuestionList(workbookID, []string{"q-1", "q-2"})
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, list)

	// then
	require.NoError(t, err)

	// when: reload
	loaded, err := repo.FindByWorkbookID(ctx, workbookID)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, loaded.Size())
	assert.True(t, loaded.Contains("q-1"))
	assert.True(t, loaded.Contains("q-2"))
	assert.Equal(t, 1, loaded.Version())
}

func Test_ActiveQuestionListRepository_Save_shouldReturnConcurrentModification_whenVersionMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-concurrent-" + t.Name()

	first, err := domain.NewActiveQuestionList(workbookID, []string{"q-1"})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, first))

	// stale aggregate that still believes the doc is at version 0
	stale, err := domain.NewActiveQuestionList(workbookID, []string{"q-1", "q-2"})
	require.NoError(t, err)

	// when
	err = repo.Save(ctx, stale)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_ActiveQuestionListRepository_Save_shouldReturnErrActiveQuestionListNotFound_whenDocWasDeletedAfterLoad(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a saved active question list is deleted out-of-band before a stale aggregate tries to save
	client := setupFirestoreClient(t)
	repo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-deleted-then-save-" + t.Name()

	initial, err := domain.NewActiveQuestionList(workbookID, []string{"q-1"})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, initial))

	loaded, err := repo.FindByWorkbookID(ctx, workbookID)
	require.NoError(t, err)

	// delete the underlying document directly via the firestore client
	_, err = client.Collection(activeQuestionListsCollectionForTest).Doc(workbookID).Delete(ctx)
	require.NoError(t, err)

	// when: the stale loaded aggregate tries to save
	require.NoError(t, loaded.Add("q-2"))
	err = repo.Save(ctx, loaded)

	// then: callers see a domain not-found, not a generic error
	require.ErrorIs(t, err, domain.ErrActiveQuestionListNotFound)
	assert.NotErrorIs(t, err, libversioned.ErrConcurrentModification,
		"deleted document must surface as not-found, not as concurrent modification")
}
