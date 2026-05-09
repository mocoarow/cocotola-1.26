package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const (
	fixtureAudioInputHashA = "a1b2c3d4e5f60718293a4b5c6d7e8f9001020304050607080910111213141516"
	fixtureAudioInputHashB = "b2c3d4e5f60718293a4b5c6d7e8f9001020304050607080910111213141516a1"
)

func newQuestionWithAudio(t *testing.T, repo *gateway.QuestionRepository, workbookID string, status domainquestion.AudioGenerationStatus, inputHash string, audioUpdatedAt time.Time) *domainquestion.Question {
	t.Helper()
	ctx := context.Background()

	id, err := uuid.NewV7()
	require.NoError(t, err)
	now := time.Now()
	q, err := domainquestion.NewQuestion(id.String(), workbookID, domainquestion.TypeWordFill(), fixtureWordFillContentForGateway, []string{"lang:en"}, 0, now, now)
	require.NoError(t, err)

	ag, err := domainquestion.NewAudioGeneration(status, inputHash, nil, audioUpdatedAt, 0, "")
	require.NoError(t, err)
	q.SetAudioGeneration(ag)

	require.NoError(t, repo.Save(ctx, q))
	return q
}

func Test_QuestionRepository_FindPendingAudio_shouldReturnOnlyPendingItems(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	now := time.Now()
	workbookID := "test-wb-pending-" + uuid.NewString()
	pending := newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now)
	_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHashB, now)
	_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusReady(), fixtureAudioInputHashA, now)
	_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusFailed(), fixtureAudioInputHashB, now)

	// when
	got, err := repo.FindPendingAudio(ctx, 50)

	// then
	require.NoError(t, err)

	// Filter to this test's workbook (the collection-group query is global).
	var inWorkbook []domainquestion.Question
	for _, q := range got {
		if q.WorkbookID() == workbookID {
			inWorkbook = append(inWorkbook, q)
		}
	}
	require.Len(t, inWorkbook, 1)
	assert.Equal(t, pending.ID(), inWorkbook[0].ID())
	assert.Equal(t, workbookID, inWorkbook[0].WorkbookID())
	require.NotNil(t, inWorkbook[0].AudioGeneration())
	assert.Equal(t, "pending", inWorkbook[0].AudioGeneration().Status().Value())
}

func Test_QuestionRepository_FindPendingAudio_shouldOrderByUpdatedAtAscending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: three pending items at different audio updatedAt timestamps
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-pending-order-" + uuid.NewString()
	now := time.Now()
	oldest := newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now.Add(-3*time.Hour))
	middle := newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now.Add(-2*time.Hour))
	newest := newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now.Add(-time.Hour))

	// when
	got, err := repo.FindPendingAudio(ctx, 50)

	// then
	require.NoError(t, err)
	indexOf := func(qs []domainquestion.Question, id string) int {
		for i, q := range qs {
			if q.ID() == id {
				return i
			}
		}
		return -1
	}
	iOld := indexOf(got, oldest.ID())
	iMid := indexOf(got, middle.ID())
	iNew := indexOf(got, newest.ID())
	require.NotEqual(t, -1, iOld)
	require.NotEqual(t, -1, iMid)
	require.NotEqual(t, -1, iNew)
	assert.Less(t, iOld, iMid, "oldest entry must precede middle")
	assert.Less(t, iMid, iNew, "middle entry must precede newest")
}

func Test_QuestionRepository_FindPendingAudio_shouldRespectLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: 3 pending items
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	workbookID := "test-wb-pending-limit-" + uuid.NewString()
	now := time.Now()
	for i := range 3 {
		_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now.Add(time.Duration(i)*time.Second))
	}

	// when
	got, err := repo.FindPendingAudio(ctx, 2)

	// then
	require.NoError(t, err)
	assert.LessOrEqual(t, len(got), 2)
}

func Test_QuestionRepository_FindPendingAudio_shouldReturnNil_whenLimitIsZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)

	// when
	got, err := repo.FindPendingAudio(ctx, 0)

	// then
	require.NoError(t, err)
	assert.Nil(t, got)
}

func Test_QuestionRepository_FindPendingAudio_shouldSpanMultipleWorkbooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: pending items in two different workbooks
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	now := time.Now()
	workbookA := "test-wb-pending-cross-a-" + uuid.NewString()
	workbookB := "test-wb-pending-cross-b-" + uuid.NewString()
	qa := newQuestionWithAudio(t, repo, workbookA, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now.Add(-2*time.Hour))
	qb := newQuestionWithAudio(t, repo, workbookB, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashB, now.Add(-time.Hour))

	// when
	got, err := repo.FindPendingAudio(ctx, 50)

	// then
	require.NoError(t, err)
	foundA, foundB := false, false
	for _, q := range got {
		if q.ID() == qa.ID() {
			foundA = true
			assert.Equal(t, workbookA, q.WorkbookID(), "workbookID must be reconstructed from parent path")
		}
		if q.ID() == qb.ID() {
			foundB = true
			assert.Equal(t, workbookB, q.WorkbookID(), "workbookID must be reconstructed from parent path")
		}
	}
	assert.True(t, foundA, "expected pending question from workbook A in result")
	assert.True(t, foundB, "expected pending question from workbook B in result")
}

func Test_QuestionRepository_FindStaleGenerating_shouldReturnOnlyGeneratingItemsOlderThanStaleBefore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)
	now := time.Now()
	staleBefore := now.Add(-30 * time.Minute)
	workbookID := "test-wb-stale-" + uuid.NewString()
	stale := newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHashA, now.Add(-2*time.Hour))
	_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusGenerating(), fixtureAudioInputHashB, now.Add(-5*time.Minute))
	_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusPending(), fixtureAudioInputHashA, now.Add(-3*time.Hour))
	_ = newQuestionWithAudio(t, repo, workbookID, domainquestion.AudioGenerationStatusReady(), fixtureAudioInputHashA, now.Add(-3*time.Hour))

	// when
	got, err := repo.FindStaleGenerating(ctx, staleBefore, 50)

	// then
	require.NoError(t, err)
	var inWorkbook []domainquestion.Question
	for _, q := range got {
		if q.WorkbookID() == workbookID {
			inWorkbook = append(inWorkbook, q)
		}
	}
	require.Len(t, inWorkbook, 1)
	assert.Equal(t, stale.ID(), inWorkbook[0].ID())
}

func Test_QuestionRepository_FindStaleGenerating_shouldReturnNil_whenLimitIsZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewQuestionRepository(client)

	// when
	got, err := repo.FindStaleGenerating(ctx, time.Now(), 0)

	// then
	require.NoError(t, err)
	assert.Nil(t, got)
}
