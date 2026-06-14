package study_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
)

func Test_NewDailyStat_shouldReturnDailyStat_whenValid(t *testing.T) {
	t.Parallel()

	// when
	stat, err := study.NewDailyStat("2026-06-14", 10, 7)

	// then
	require.NoError(t, err)
	assert.Equal(t, "2026-06-14", stat.Date)
	assert.Equal(t, 10, stat.AnsweredCount)
	assert.Equal(t, 7, stat.CorrectCount)
}

func Test_NewDailyStat_shouldReturnError_whenDateIsMalformed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		date string
	}{
		{name: "empty", date: ""},
		{name: "missingZeroPad", date: "2026-6-14"},
		{name: "slashes", date: "2026/06/14"},
		{name: "isoTimestamp", date: "2026-06-14T00:00:00Z"},
		{name: "garbage", date: "not-a-date"},
		{name: "outOfRange", date: "2026-13-40"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			_, err := study.NewDailyStat(tt.date, 1, 0)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewDailyStat_shouldReturnError_whenAnsweredCountIsNegative(t *testing.T) {
	t.Parallel()

	// when
	_, err := study.NewDailyStat("2026-06-14", -1, 0)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewDailyStat_shouldReturnError_whenCorrectCountIsNegative(t *testing.T) {
	t.Parallel()

	// when
	_, err := study.NewDailyStat("2026-06-14", 5, -1)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewDailyStat_shouldReturnError_whenCorrectExceedsAnswered(t *testing.T) {
	t.Parallel()

	// when
	_, err := study.NewDailyStat("2026-06-14", 5, 6)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_IsValidDateKey_shouldReturnTrue_whenCanonical(t *testing.T) {
	t.Parallel()

	// when
	got := study.IsValidDateKey("2026-06-14")

	// then
	assert.True(t, got)
}

func Test_IsValidDateKey_shouldReturnFalse_whenMalformed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		v    string
	}{
		{name: "empty", v: ""},
		{name: "missingZeroPad", v: "2026-6-14"},
		{name: "slashes", v: "2026/06/14"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			got := study.IsValidDateKey(tt.v)

			// then
			assert.False(t, got)
		})
	}
}
