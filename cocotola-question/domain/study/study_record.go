package study

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// StudyRecord tracks a user's learning progress for a single question.
type StudyRecord struct {
	workbookID         string
	questionID         string
	consecutiveCorrect int
	lastAnsweredAt     time.Time
	nextDueAt          time.Time
	totalCorrect       int
	totalIncorrect     int
	version            int
}

// NewStudyRecord creates a new StudyRecord for a question that has never been studied.
func NewStudyRecord(workbookID string, questionID string) (*StudyRecord, error) {
	if workbookID == "" {
		return nil, fmt.Errorf("workbook id is required: %w", domain.ErrInvalidArgument)
	}
	if questionID == "" {
		return nil, fmt.Errorf("question id is required: %w", domain.ErrInvalidArgument)
	}
	return &StudyRecord{
		workbookID: workbookID,
		questionID: questionID,
	}, nil
}

// ReconstructStudyRecord reconstitutes a StudyRecord from persistence.
func ReconstructStudyRecord(
	workbookID string,
	questionID string,
	consecutiveCorrect int,
	lastAnsweredAt time.Time,
	nextDueAt time.Time,
	totalCorrect int,
	totalIncorrect int,
) *StudyRecord {
	return &StudyRecord{
		workbookID:         workbookID,
		questionID:         questionID,
		consecutiveCorrect: consecutiveCorrect,
		lastAnsweredAt:     lastAnsweredAt,
		nextDueAt:          nextDueAt,
		totalCorrect:       totalCorrect,
		totalIncorrect:     totalIncorrect,
	}
}

// RecordCorrect records a correct answer and updates the next due date.
func (r *StudyRecord) RecordCorrect(now time.Time) {
	r.consecutiveCorrect++
	r.lastAnsweredAt = now
	r.nextDueAt = CalculateNextDue(r.consecutiveCorrect, now)
	r.totalCorrect++
}

// RecordIncorrect records an incorrect answer, resets consecutive correct count,
// and sets the next due date to 1 day later.
func (r *StudyRecord) RecordIncorrect(now time.Time) {
	r.consecutiveCorrect = 0
	r.lastAnsweredAt = now
	r.nextDueAt = now.AddDate(0, 0, 1)
	r.totalIncorrect++
}

func (r *StudyRecord) WorkbookID() string      { return r.workbookID }
func (r *StudyRecord) QuestionID() string       { return r.questionID }
func (r *StudyRecord) ConsecutiveCorrect() int  { return r.consecutiveCorrect }
func (r *StudyRecord) LastAnsweredAt() time.Time { return r.lastAnsweredAt }
func (r *StudyRecord) NextDueAt() time.Time     { return r.nextDueAt }
func (r *StudyRecord) TotalCorrect() int        { return r.totalCorrect }
func (r *StudyRecord) TotalIncorrect() int      { return r.totalIncorrect }

// Version returns the persisted version (0 = new, not yet saved).
func (r *StudyRecord) Version() int { return r.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
// This method is intended for use by the persistence layer only.
func (r *StudyRecord) SetVersion(version int) { r.version = version }
