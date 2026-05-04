package study_test

import (
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRecord_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := study.NewRecord("", "q1")

	// then
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRecord_shouldReturnError_whenQuestionIDIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := study.NewRecord("wb1", "")

	// then
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRecord_shouldReturnRecord_whenValid(t *testing.T) {
	t.Parallel()

	// when
	record, err := study.NewRecord("wb1", "q1")

	// then
	require.NoError(t, err)
	assert.Equal(t, "wb1", record.WorkbookID())
	assert.Equal(t, "q1", record.QuestionID())
	assert.Equal(t, 0, record.ConsecutiveCorrect())
	assert.Equal(t, 0, record.TotalCorrect())
	assert.Equal(t, 0, record.TotalIncorrect())
}

func Test_Record_RecordCorrect_shouldIncrementConsecutiveCorrect(t *testing.T) {
	t.Parallel()

	// given
	record, err := study.NewRecord("wb1", "q1")
	require.NoError(t, err)
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	// when
	record.RecordCorrect(now)

	// then
	assert.Equal(t, 1, record.ConsecutiveCorrect())
	assert.Equal(t, 1, record.TotalCorrect())
	assert.Equal(t, 0, record.TotalIncorrect())
	assert.Equal(t, now, record.LastAnsweredAt())
	assert.Equal(t, now.AddDate(0, 0, 1), record.NextDueAt())
}

func Test_Record_RecordCorrect_shouldDoubleInterval_whenConsecutive(t *testing.T) {
	t.Parallel()

	// given
	record, err := study.NewRecord("wb1", "q1")
	require.NoError(t, err)
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	// when
	record.RecordCorrect(now)
	record.RecordCorrect(now)
	record.RecordCorrect(now)

	// then
	assert.Equal(t, 3, record.ConsecutiveCorrect())
	assert.Equal(t, 3, record.TotalCorrect())
	assert.Equal(t, now.AddDate(0, 0, 7), record.NextDueAt())
}

func Test_Record_RecordIncorrect_shouldResetConsecutiveCorrect(t *testing.T) {
	t.Parallel()

	// given
	record, err := study.NewRecord("wb1", "q1")
	require.NoError(t, err)
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	record.RecordCorrect(now)
	record.RecordCorrect(now)

	// when
	record.RecordIncorrect(now)

	// then
	assert.Equal(t, 0, record.ConsecutiveCorrect())
	assert.Equal(t, 2, record.TotalCorrect())
	assert.Equal(t, 1, record.TotalIncorrect())
	assert.Equal(t, now.Add(study.IncorrectRetryDelay), record.NextDueAt())
}

func Test_Record_RecordCorrectAfterIncorrect_shouldRestartInterval(t *testing.T) {
	t.Parallel()

	// given
	record, err := study.NewRecord("wb1", "q1")
	require.NoError(t, err)
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	record.RecordCorrect(now)
	record.RecordCorrect(now)
	record.RecordIncorrect(now)

	// when
	record.RecordCorrect(now)

	// then
	assert.Equal(t, 1, record.ConsecutiveCorrect())
	assert.Equal(t, now.AddDate(0, 0, 1), record.NextDueAt())
}
