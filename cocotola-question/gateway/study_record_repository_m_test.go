package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

func Test_StudyRecordRepository_Save_shouldPersistRecord(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-save-" + t.Name()
	workbookID := "wb-1"
	questionID := "q-1"
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	nextDue := now.Add(24 * time.Hour)

	record := study.ReconstructStudyRecord(workbookID, questionID, 1, now, nextDue, 1, 0)

	// when
	err := repo.Save(ctx, userID, record)

	// then
	require.NoError(t, err)

	// when - reload
	loaded, err := repo.FindByID(ctx, userID, workbookID, questionID)

	// then
	require.NoError(t, err)
	assert.Equal(t, workbookID, loaded.WorkbookID())
	assert.Equal(t, questionID, loaded.QuestionID())
	assert.Equal(t, 1, loaded.ConsecutiveCorrect())
	assert.Equal(t, now.UTC(), loaded.LastAnsweredAt().UTC())
	assert.Equal(t, nextDue.UTC(), loaded.NextDueAt().UTC())
	assert.Equal(t, 1, loaded.TotalCorrect())
	assert.Equal(t, 0, loaded.TotalIncorrect())
}

func Test_StudyRecordRepository_Save_shouldUpdateExistingRecord(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-update-" + t.Name()
	workbookID := "wb-1"
	questionID := "q-1"
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	record1 := study.ReconstructStudyRecord(workbookID, questionID, 1, now, now.Add(24*time.Hour), 1, 0)
	require.NoError(t, repo.Save(ctx, userID, record1))

	// when - update with new values (use version from first save)
	later := now.Add(1 * time.Hour)
	record2 := study.ReconstructStudyRecord(workbookID, questionID, 2, later, later.Add(48*time.Hour), 2, 0)
	record2.SetVersion(record1.Version())
	err := repo.Save(ctx, userID, record2)

	// then
	require.NoError(t, err)

	loaded, err := repo.FindByID(ctx, userID, workbookID, questionID)
	require.NoError(t, err)
	assert.Equal(t, 2, loaded.ConsecutiveCorrect())
	assert.Equal(t, 2, loaded.TotalCorrect())
}

func Test_StudyRecordRepository_FindByID_shouldReturnNotFound_whenRecordNotExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-notfound-" + t.Name()

	// when
	_, err := repo.FindByID(ctx, userID, "wb-nonexistent", "q-nonexistent")

	// then
	require.ErrorIs(t, err, domain.ErrStudyRecordNotFound)
}

func Test_StudyRecordRepository_Save_shouldReturnError_whenVersionConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-conflict-" + t.Name()
	workbookID := "wb-1"
	questionID := "q-1"
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	record := study.ReconstructStudyRecord(workbookID, questionID, 1, now, now.Add(24*time.Hour), 1, 0)
	require.NoError(t, repo.Save(ctx, userID, record))

	// when - attempt to save with stale version (0 instead of 1)
	staleRecord := study.ReconstructStudyRecord(workbookID, questionID, 2, now, now.Add(48*time.Hour), 2, 0)
	err := repo.Save(ctx, userID, staleRecord)

	// then
	require.ErrorIs(t, err, domain.ErrConcurrentModification)
}

func Test_StudyRecordRepository_FindByWorkbookID_shouldReturnEmpty_whenNoRecords(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-empty-" + t.Name()

	// when
	records, err := repo.FindByWorkbookID(ctx, userID, "wb-nonexistent")

	// then
	require.NoError(t, err)
	assert.Len(t, records, 0)
}

func Test_StudyRecordRepository_FindByWorkbookID_shouldReturnRecords_whenRecordsExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-list-" + t.Name()
	workbookID := "wb-1"
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	record1 := study.ReconstructStudyRecord(workbookID, "q-1", 1, now, now.Add(24*time.Hour), 1, 0)
	record2 := study.ReconstructStudyRecord(workbookID, "q-2", 0, now, now.Add(10*time.Minute), 0, 1)
	require.NoError(t, repo.Save(ctx, userID, record1))
	require.NoError(t, repo.Save(ctx, userID, record2))

	// when
	records, err := repo.FindByWorkbookID(ctx, userID, workbookID)

	// then
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func Test_StudyRecordRepository_FindByWorkbookID_shouldNotReturnRecords_whenDifferentWorkbook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyRecordRepository(client)
	userID := "test-user-filter-" + t.Name()
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	record := study.ReconstructStudyRecord("wb-A", "q-1", 1, now, now.Add(24*time.Hour), 1, 0)
	require.NoError(t, repo.Save(ctx, userID, record))

	// when
	records, err := repo.FindByWorkbookID(ctx, userID, "wb-B")

	// then
	require.NoError(t, err)
	assert.Len(t, records, 0)
}
