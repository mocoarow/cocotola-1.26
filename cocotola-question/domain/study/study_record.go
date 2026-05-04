package study

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// Record tracks a user's learning progress for a single question.
type Record struct {
	workbookID         string
	questionID         string
	consecutiveCorrect int
	lastAnsweredAt     time.Time
	nextDueAt          time.Time
	totalCorrect       int
	totalIncorrect     int
	version            int
}

// NewRecord creates a new Record for a question that has never been studied.
func NewRecord(workbookID string, questionID string) (*Record, error) {
	if workbookID == "" {
		return nil, fmt.Errorf("workbook id is required: %w", domain.ErrInvalidArgument)
	}
	if questionID == "" {
		return nil, fmt.Errorf("question id is required: %w", domain.ErrInvalidArgument)
	}
	return &Record{
		workbookID:         workbookID,
		questionID:         questionID,
		consecutiveCorrect: 0,
		lastAnsweredAt:     time.Time{},
		nextDueAt:          time.Time{},
		totalCorrect:       0,
		totalIncorrect:     0,
		version:            0,
	}, nil
}

// ReconstructRecord reconstitutes a Record from persistence.
func ReconstructRecord(
	workbookID string,
	questionID string,
	consecutiveCorrect int,
	lastAnsweredAt time.Time,
	nextDueAt time.Time,
	totalCorrect int,
	totalIncorrect int,
) *Record {
	return &Record{
		workbookID:         workbookID,
		questionID:         questionID,
		consecutiveCorrect: consecutiveCorrect,
		lastAnsweredAt:     lastAnsweredAt,
		nextDueAt:          nextDueAt,
		totalCorrect:       totalCorrect,
		totalIncorrect:     totalIncorrect,
		version:            0,
	}
}

// RecordCorrect records a correct answer and updates the next due date.
func (r *Record) RecordCorrect(now time.Time) {
	r.consecutiveCorrect++
	r.lastAnsweredAt = now
	r.nextDueAt = CalculateNextDue(r.consecutiveCorrect, now)
	r.totalCorrect++
}

// IncorrectRetryDelay is how long the SRS waits before re-presenting a
// question the user just got wrong. It's intentionally short — wrong answers
// are "not yet mastered" and should resurface in the same study session (or
// the next time the user re-opens the workbook), unlike correctly-answered
// questions whose next-due grows on the standard SRS curve.
const IncorrectRetryDelay = 5 * time.Minute

// RecordIncorrect records an incorrect answer, resets consecutive correct count,
// and schedules the next presentation IncorrectRetryDelay later.
func (r *Record) RecordIncorrect(now time.Time) {
	r.consecutiveCorrect = 0
	r.lastAnsweredAt = now
	r.nextDueAt = now.Add(IncorrectRetryDelay)
	r.totalIncorrect++
}

// WorkbookID returns the workbook that owns this record.
func (r *Record) WorkbookID() string { return r.workbookID }

// QuestionID returns the question this record tracks.
func (r *Record) QuestionID() string { return r.questionID }

// ConsecutiveCorrect returns the current streak of correct answers.
func (r *Record) ConsecutiveCorrect() int { return r.consecutiveCorrect }

// LastAnsweredAt returns when the question was last answered.
func (r *Record) LastAnsweredAt() time.Time { return r.lastAnsweredAt }

// NextDueAt returns when the question is next due for review.
func (r *Record) NextDueAt() time.Time { return r.nextDueAt }

// TotalCorrect returns the total number of correct answers.
func (r *Record) TotalCorrect() int { return r.totalCorrect }

// TotalIncorrect returns the total number of incorrect answers.
func (r *Record) TotalIncorrect() int { return r.totalIncorrect }

// Version returns the persisted version (0 = new, not yet saved).
func (r *Record) Version() int { return r.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
// This method is intended for use by the persistence layer only.
func (r *Record) SetVersion(version int) { r.version = version }
