package gateway_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const fixtureExistingActiveQuestionID = "q-existing"

func saveActiveQuestionList(t *testing.T, repo *gateway.ActiveQuestionListRepository, workbookID string, questionIDs ...string) {
	t.Helper()
	list, err := domain.NewActiveQuestionList(workbookID, questionIDs)
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), list))
}

func findActiveQuestionList(t *testing.T, repo *gateway.ActiveQuestionListRepository, workbookID string) *domain.ActiveQuestionList {
	t.Helper()
	list, err := repo.FindByWorkbookID(context.Background(), workbookID)
	require.NoError(t, err)
	return list
}

// savedQuestionAsNew persists a question and returns it reset to version 0,
// so that Create collides with the stored document.
func savedQuestionAsNew(t *testing.T, repo *gateway.QuestionRepository, workbookID string) *domainquestion.Question {
	t.Helper()
	q := newQuestion(t, workbookID, 0)
	require.NoError(t, repo.Save(context.Background(), q))
	q.SetVersion(0)
	return q
}

func Test_QuestionRepository_Create_shouldAddQuestionIDToActiveList_whenListExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	listRepo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-create-add-" + t.Name()
	saveActiveQuestionList(t, listRepo, workbookID, fixtureExistingActiveQuestionID)
	q := newQuestion(t, workbookID, 0)

	// when
	err := repo.Create(ctx, q)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{fixtureExistingActiveQuestionID, q.ID()}, findActiveQuestionList(t, listRepo, workbookID).Entries())
}

func Test_QuestionRepository_Create_shouldIncrementActiveListVersion_whenListExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	listRepo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-create-version-" + t.Name()
	saveActiveQuestionList(t, listRepo, workbookID, fixtureExistingActiveQuestionID)

	// when
	err := repo.Create(ctx, newQuestion(t, workbookID, 0))

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, findActiveQuestionList(t, listRepo, workbookID).Version())
}

func Test_QuestionRepository_Create_shouldCreateActiveList_whenListAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	listRepo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-create-nolist-" + t.Name()
	q := newQuestion(t, workbookID, 0)

	// when
	err := repo.Create(ctx, q)

	// then
	require.NoError(t, err)
	assert.Equal(t, []string{q.ID()}, findActiveQuestionList(t, listRepo, workbookID).Entries())
}

func Test_QuestionRepository_Create_shouldPersistQuestion_whenCreated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-create-persist-" + t.Name()
	q := newQuestion(t, workbookID, 3)

	// when
	err := repo.Create(ctx, q)

	// then
	require.NoError(t, err)
	found, err := repo.FindByID(ctx, workbookID, q.ID())
	require.NoError(t, err)
	assert.Equal(t, 3, found.OrderIndex())
}

func Test_QuestionRepository_Create_shouldPersistVersionOne_whenCreated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-create-persist-version-" + t.Name()
	q := newQuestion(t, workbookID, 0)

	// when
	err := repo.Create(ctx, q)

	// then
	require.NoError(t, err)
	found, err := repo.FindByID(ctx, workbookID, q.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, found.Version())
}

func Test_QuestionRepository_Create_shouldSetVersionToOne_whenCreated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	q := newQuestion(t, "test-wb-create-setversion-"+t.Name(), 0)

	// when
	err := repo.Create(ctx, q)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, q.Version())
}

func Test_QuestionRepository_Create_shouldReturnInvalidArgument_whenVersionIsNotZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	q := newQuestion(t, "test-wb-create-loaded-"+t.Name(), 0)
	q.SetVersion(1)

	// when
	err := repo.Create(ctx, q)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_QuestionRepository_Create_shouldReturnConcurrentModification_whenQuestionAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	q := savedQuestionAsNew(t, repo, "test-wb-create-dup-"+t.Name())

	// when
	err := repo.Create(ctx, q)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_QuestionRepository_Create_shouldNotChangeActiveList_whenQuestionAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	listRepo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-create-dup-list-" + t.Name()
	saveActiveQuestionList(t, listRepo, workbookID, fixtureExistingActiveQuestionID)
	q := savedQuestionAsNew(t, repo, workbookID)

	// when
	_ = repo.Create(ctx, q)

	// then
	assert.Equal(t, []string{fixtureExistingActiveQuestionID}, findActiveQuestionList(t, listRepo, workbookID).Entries())
}

func Test_QuestionRepository_Create_shouldKeepActiveListMatchingQuestions_whenCreatedConcurrently(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	const concurrency = 8
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	listRepo := gateway.NewActiveQuestionListRepository(client)
	workbookID := "test-wb-create-concurrent-" + t.Name()
	questions := make([]*domainquestion.Question, concurrency)
	for i := range questions {
		questions[i] = newQuestion(t, workbookID, i)
	}

	// when
	var wg sync.WaitGroup
	for _, q := range questions {
		wg.Go(func() {
			if err := repo.Create(ctx, q); err != nil {
				t.Logf("create question %s: %v", q.ID(), err)
			}
		})
	}
	wg.Wait()

	// then
	stored, err := repo.FindByWorkbookID(ctx, workbookID)
	require.NoError(t, err)
	require.NotEmpty(t, stored)
	assert.ElementsMatch(t, questionIDs(stored), findActiveQuestionList(t, listRepo, workbookID).Entries())
}
