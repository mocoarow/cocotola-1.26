package study

import "time"

const maxIntervalDays = 180

// CalculateNextDue returns the next due time based on consecutive correct answers.
func CalculateNextDue(consecutiveCorrect int, now time.Time) time.Time {
	days := intervalDays(consecutiveCorrect)
	return now.AddDate(0, 0, days)
}

// intervalDays returns the number of days until the next review based on
// consecutive correct answers. Uses a predefined table for counts 1-4
// (1, 3, 7, 14 days), then doubles from the last table entry for higher
// counts, capped at maxIntervalDays (180 days).
func intervalDays(consecutiveCorrect int) int {
	// intervals maps consecutive correct count to next interval in days.
	// For counts >= 5, the interval doubles from the previous one.
	intervals := []int{1, 3, 7, 14}

	if consecutiveCorrect <= 0 {
		return 1
	}
	idx := consecutiveCorrect - 1
	if idx < len(intervals) {
		return intervals[idx]
	}

	days := intervals[len(intervals)-1]
	for i := len(intervals); i < consecutiveCorrect; i++ {
		days *= 2
		if days > maxIntervalDays {
			return maxIntervalDays
		}
	}
	return days
}
