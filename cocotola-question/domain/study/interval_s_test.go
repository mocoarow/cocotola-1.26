package study_test

import (
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	"github.com/stretchr/testify/assert"
)

func Test_CalculateNextDue_shouldReturnCorrectInterval(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		consecutiveCorrect int
		wantDays           int
	}{
		{name: "first_correct", consecutiveCorrect: 1, wantDays: 1},
		{name: "second_correct", consecutiveCorrect: 2, wantDays: 3},
		{name: "third_correct", consecutiveCorrect: 3, wantDays: 7},
		{name: "fourth_correct", consecutiveCorrect: 4, wantDays: 14},
		{name: "fifth_correct", consecutiveCorrect: 5, wantDays: 28},
		{name: "sixth_correct", consecutiveCorrect: 6, wantDays: 56},
		{name: "seventh_correct", consecutiveCorrect: 7, wantDays: 112},
		{name: "eighth_correct_capped", consecutiveCorrect: 8, wantDays: 180},
		{name: "tenth_correct_capped", consecutiveCorrect: 10, wantDays: 180},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			result := study.CalculateNextDue(tt.consecutiveCorrect, now)

			// then
			expected := now.AddDate(0, 0, tt.wantDays)
			assert.Equal(t, expected, result)
		})
	}
}

func Test_CalculateNextDue_shouldReturn1Day_whenZeroOrNegative(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		consecutiveCorrect int
	}{
		{name: "zero", consecutiveCorrect: 0},
		{name: "negative", consecutiveCorrect: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			result := study.CalculateNextDue(tt.consecutiveCorrect, now)

			// then
			expected := now.AddDate(0, 0, 1)
			assert.Equal(t, expected, result)
		})
	}
}
