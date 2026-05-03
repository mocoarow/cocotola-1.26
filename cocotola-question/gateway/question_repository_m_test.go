package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const fixtureWordFillContentForGateway = `{"source":{"text":"apple","lang":"en"},"target":{"text":"{{りんご}}","lang":"ja"}}`

func newQuestion(t *testing.T, workbookID string, orderIndex int) *domainquestion.Question {
	t.Helper()
	id, err := uuid.NewV7()
	require.NoError(t, err)
	now := time.Now()
	q, err := domainquestion.NewQuestion(id.String(), workbookID, domainquestion.TypeWordFill(), fixtureWordFillContentForGateway, []string{"lang:en"}, orderIndex, now, now)
	require.NoError(t, err)
	return q
}

func Test_QuestionRepository_Save_shouldInsertAndIncrementVersion_whenVersionIsZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-insert-" + t.Name()
	q := newQuestion(t, workbookID, 0)

	// when
	err := repo.Save(ctx, q)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, q.Version())
}

func Test_QuestionRepository_Save_shouldPersistAggregate_whenInsertSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-persist-" + t.Name()
	q := newQuestion(t, workbookID, 3)
	require.NoError(t, repo.Save(ctx, q))

	// when
	loaded, err := repo.FindByID(ctx, workbookID, q.ID())

	// then
	require.NoError(t, err)
	assert.Equal(t, q.ID(), loaded.ID())
	assert.Equal(t, workbookID, loaded.WorkbookID())
	assert.Equal(t, "word_fill", loaded.QuestionType().Value())
	assert.JSONEq(t, fixtureWordFillContentForGateway, loaded.Content())
	assert.Equal(t, []string{"lang:en"}, loaded.Tags())
	assert.Equal(t, 3, loaded.OrderIndex())
	assert.Equal(t, 1, loaded.Version())
}

func Test_QuestionRepository_Save_shouldUpdateAndBumpVersion_whenVersionMatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-update-" + t.Name()
	q := newQuestion(t, workbookID, 0)
	require.NoError(t, repo.Save(ctx, q))

	updatedContent := `{"source":{"text":"banana","lang":"en"},"target":{"text":"{{バナナ}}","lang":"ja"}}`
	require.NoError(t, q.Edit(updatedContent, []string{"lang:ja"}, 5, time.Now()))

	// when
	err := repo.Save(ctx, q)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, q.Version())

	loaded, err := repo.FindByID(ctx, workbookID, q.ID())
	require.NoError(t, err)
	assert.JSONEq(t, updatedContent, loaded.Content())
	assert.Equal(t, []string{"lang:ja"}, loaded.Tags())
	assert.Equal(t, 5, loaded.OrderIndex())
	assert.Equal(t, 2, loaded.Version())
}

func Test_QuestionRepository_Save_shouldReturnConcurrentModification_whenVersionMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-conflict-" + t.Name()
	q := newQuestion(t, workbookID, 0)
	require.NoError(t, repo.Save(ctx, q))

	// Reload, then a stale aggregate (still version 1) tries to save after another update.
	stale, err := repo.FindByID(ctx, workbookID, q.ID())
	require.NoError(t, err)
	require.NoError(t, q.Edit(fixtureWordFillContentForGateway, nil, 9, time.Now()))
	require.NoError(t, repo.Save(ctx, q)) // q is now version 2; stale still at 1

	require.NoError(t, stale.Edit(fixtureWordFillContentForGateway, nil, 99, time.Now()))

	// when
	err = repo.Save(ctx, stale)

	// then
	require.ErrorIs(t, err, domain.ErrConcurrentModification)
}

func Test_QuestionRepository_FindByWorkbookID_shouldReturnQuestionsInOrderIndexOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-order-" + t.Name()
	for _, idx := range []int{2, 0, 1} {
		require.NoError(t, repo.Save(ctx, newQuestion(t, workbookID, idx)))
	}

	// when
	got, err := repo.FindByWorkbookID(ctx, workbookID)

	// then
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, 0, got[0].OrderIndex())
	assert.Equal(t, 1, got[1].OrderIndex())
	assert.Equal(t, 2, got[2].OrderIndex())
}

func Test_QuestionRepository_FindByID_shouldReturnErrQuestionNotFound_whenMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)

	// when
	_, err := repo.FindByID(ctx, "missing-wb", "missing-q")

	// then
	require.ErrorIs(t, err, domain.ErrQuestionNotFound)
}

func Test_QuestionRepository_Delete_shouldRemoveDocument(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-delete-" + t.Name()
	q := newQuestion(t, workbookID, 0)
	require.NoError(t, repo.Save(ctx, q))

	// when
	err := repo.Delete(ctx, workbookID, q.ID())

	// then
	require.NoError(t, err)
	_, err = repo.FindByID(ctx, workbookID, q.ID())
	require.ErrorIs(t, err, domain.ErrQuestionNotFound)
}
