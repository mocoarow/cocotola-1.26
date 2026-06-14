package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

func Test_StudyDailyStatsRepository_IncrementToday_shouldCreateBucket_whenFirstWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)
	userID := "test-user-incr-create-" + t.Name()
	dateKey := "2026-06-14"
	now := time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC)

	// when
	err := repo.IncrementToday(ctx, userID, dateKey, "Asia/Tokyo", true, now)

	// then
	require.NoError(t, err)

	stats, err := repo.FindRange(ctx, userID, dateKey, dateKey)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, dateKey, stats[0].Date)
	assert.Equal(t, 1, stats[0].AnsweredCount)
	assert.Equal(t, 1, stats[0].CorrectCount)
}

func Test_StudyDailyStatsRepository_IncrementToday_shouldIncrementAnsweredOnly_whenIncorrect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)
	userID := "test-user-incr-incorrect-" + t.Name()
	dateKey := "2026-06-14"
	now := time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC)

	// when
	err := repo.IncrementToday(ctx, userID, dateKey, "Asia/Tokyo", false, now)

	// then
	require.NoError(t, err)

	stats, err := repo.FindRange(ctx, userID, dateKey, dateKey)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, 1, stats[0].AnsweredCount)
	assert.Equal(t, 0, stats[0].CorrectCount)
}

func Test_StudyDailyStatsRepository_IncrementToday_shouldAccumulate_whenMultipleWrites(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)
	userID := "test-user-incr-accumulate-" + t.Name()
	dateKey := "2026-06-14"
	now := time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC)

	// when
	require.NoError(t, repo.IncrementToday(ctx, userID, dateKey, "Asia/Tokyo", true, now))
	require.NoError(t, repo.IncrementToday(ctx, userID, dateKey, "Asia/Tokyo", false, now))
	require.NoError(t, repo.IncrementToday(ctx, userID, dateKey, "Asia/Tokyo", true, now))

	// then
	stats, err := repo.FindRange(ctx, userID, dateKey, dateKey)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, 3, stats[0].AnsweredCount)
	assert.Equal(t, 2, stats[0].CorrectCount)
}

func Test_StudyDailyStatsRepository_IncrementToday_shouldReturnError_whenInputInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)
	now := time.Now()

	tests := []struct {
		name     string
		userID   string
		dateKey  string
		timezone string
	}{
		{name: "emptyUserID", userID: "", dateKey: "2026-06-14", timezone: "Asia/Tokyo"},
		{name: "invalidDateKey", userID: "u", dateKey: "2026/06/14", timezone: "Asia/Tokyo"},
		{name: "emptyTimezone", userID: "u", dateKey: "2026-06-14", timezone: ""},
		{name: "nonexistentTimezone", userID: "u", dateKey: "2026-06-14", timezone: "Not/AZone"},
		{name: "garbageTimezone", userID: "u", dateKey: "2026-06-14", timezone: "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := repo.IncrementToday(ctx, tt.userID, tt.dateKey, tt.timezone, true, now)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_StudyDailyStatsRepository_FindRange_shouldReturnEmpty_whenNoBuckets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)
	userID := "test-user-empty-" + t.Name()

	// when
	stats, err := repo.FindRange(ctx, userID, "2026-01-01", "2026-12-31")

	// then
	require.NoError(t, err)
	assert.Empty(t, stats)
}

func Test_StudyDailyStatsRepository_FindRange_shouldReturnOnlyBucketsInRange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)
	userID := "test-user-range-" + t.Name()
	now := time.Now()

	for _, d := range []string{"2026-06-12", "2026-06-13", "2026-06-14", "2026-06-15"} {
		require.NoError(t, repo.IncrementToday(ctx, userID, d, "Asia/Tokyo", true, now))
	}

	// when
	stats, err := repo.FindRange(ctx, userID, "2026-06-13", "2026-06-14")

	// then
	require.NoError(t, err)
	require.Len(t, stats, 2)
	assert.Equal(t, "2026-06-13", stats[0].Date)
	assert.Equal(t, "2026-06-14", stats[1].Date)
}

func Test_StudyDailyStatsRepository_FindRange_shouldReturnError_whenInputInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewStudyDailyStatsRepository(client)

	tests := []struct {
		name string
		userID,
		fromKey,
		toKey string
	}{
		{name: "emptyUserID", userID: "", fromKey: "2026-06-01", toKey: "2026-06-30"},
		{name: "invalidFrom", userID: "u", fromKey: "06-01", toKey: "2026-06-30"},
		{name: "invalidTo", userID: "u", fromKey: "2026-06-01", toKey: "later"},
		{name: "reversedRange", userID: "u", fromKey: "2026-06-30", toKey: "2026-06-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := repo.FindRange(ctx, tt.userID, tt.fromKey, tt.toKey)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
