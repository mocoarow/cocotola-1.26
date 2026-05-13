// Package study provides use cases for spaced repetition study.
package study

import (
	"math/rand/v2"
	"time"
)

// UsecaseConfig holds configuration for study use cases.
type UsecaseConfig struct {
	ClockFunc   func() time.Time
	ShuffleFunc func(n int, swap func(i, j int))
}

// Now returns the current time using ClockFunc if set, otherwise time.Now.
func (c UsecaseConfig) Now() time.Time {
	if c.ClockFunc != nil {
		return c.ClockFunc()
	}
	return time.Now()
}

// Shuffle shuffles elements using ShuffleFunc if set, otherwise rand.Shuffle.
func (c UsecaseConfig) Shuffle(n int, swap func(i, j int)) {
	if c.ShuffleFunc != nil {
		c.ShuffleFunc(n, swap)
		return
	}
	rand.Shuffle(n, swap)
}

// Command composes all study use cases.
type Command struct {
	*GetStudyQuestionsQuery
	*GetStudySummaryQuery
	*RecordAnswerCommand
	*DeleteStudyHistoryCommand
	*ListStudyRecordsQuery
}

// questionRepository is the union of question lookup capabilities required by
// study use cases. The concrete *gateway.QuestionRepository satisfies it.
type questionRepository interface {
	questionBatchFinder
	questionFinder
}

// studyRecordRepository is the union of study-record capabilities required by
// study use cases. The concrete *gateway.StudyRecordRepository satisfies it.
type studyRecordRepository interface {
	studyRecordFinder
	studyRecordSaver
	studyRecordDeleter
}

// NewCommand returns a new Command composing all study use cases.
func NewCommand(
	studyRecordRepo studyRecordRepository,
	activeListRepo activeQuestionListFinder,
	questionRepo questionRepository,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
	config UsecaseConfig,
) *Command {
	return &Command{
		GetStudyQuestionsQuery:    NewGetStudyQuestionsQuery(studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, config),
		GetStudySummaryQuery:      NewGetStudySummaryQuery(studyRecordRepo, activeListRepo, workbookRepo, authChecker, config),
		RecordAnswerCommand:       NewRecordAnswerCommand(studyRecordRepo, studyRecordRepo, activeListRepo, questionRepo, workbookRepo, authChecker, config),
		DeleteStudyHistoryCommand: NewDeleteStudyHistoryCommand(studyRecordRepo, workbookRepo, authChecker),
		ListStudyRecordsQuery:     NewListStudyRecordsQuery(studyRecordRepo, workbookRepo, authChecker),
	}
}
