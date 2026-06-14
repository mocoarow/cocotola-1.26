package study

import (
	"fmt"
	"regexp"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// dateKeyPattern enforces the canonical YYYY-MM-DD bucket key used both as
// document ID and as the on-the-wire date field. Anchored on both ends so any
// malformed input (missing zero-padding, slashes, ISO timestamps) is rejected
// at the boundary rather than producing silently-wrong aggregations downstream.
var dateKeyPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// IsValidDateKey reports whether the value is a canonical YYYY-MM-DD date key.
func IsValidDateKey(value string) bool {
	if !dateKeyPattern.MatchString(value) {
		return false
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return false
	}
	return true
}

// DailyStat is one day's bucket of a user's study activity, aggregated across
// every workbook they've touched. It powers the dashboard contribution graph
// and streak/goal cards.
type DailyStat struct {
	Date          string
	AnsweredCount int
	CorrectCount  int
}

// NewDailyStat returns a validated DailyStat.
func NewDailyStat(date string, answeredCount int, correctCount int) (DailyStat, error) {
	if !IsValidDateKey(date) {
		return DailyStat{}, fmt.Errorf("date must be YYYY-MM-DD: %w", domain.ErrInvalidArgument)
	}
	if answeredCount < 0 {
		return DailyStat{}, fmt.Errorf("answered count must be non-negative: %w", domain.ErrInvalidArgument)
	}
	if correctCount < 0 {
		return DailyStat{}, fmt.Errorf("correct count must be non-negative: %w", domain.ErrInvalidArgument)
	}
	if correctCount > answeredCount {
		return DailyStat{}, fmt.Errorf("correct count must not exceed answered count: %w", domain.ErrInvalidArgument)
	}
	return DailyStat{
		Date:          date,
		AnsweredCount: answeredCount,
		CorrectCount:  correctCount,
	}, nil
}
