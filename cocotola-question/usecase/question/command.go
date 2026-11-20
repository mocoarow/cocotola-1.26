// Package question provides use cases for question management.
package question

import "time"

// UsecaseConfig holds configuration for question use cases.
type UsecaseConfig struct {
	ClockFunc func() time.Time
}

// Now returns the current time using ClockFunc if set, otherwise time.Now.
func (c UsecaseConfig) Now() time.Time {
	if c.ClockFunc != nil {
		return c.ClockFunc()
	}
	return time.Now()
}

// Command composes all question management use cases.
type Command struct {
	*AddQuestionCommand
	*GetQuestionQuery
	*ListQuestionsQuery
	*UpdateQuestionCommand
	*DeleteQuestionCommand
}

// NewCommand returns a new Command composing all question use cases.
func NewCommand(
	questionSaverRepo questionSaver,
	questionFinderRepo questionFinder,
	questionDeleterRepo questionDeleter,
	workbookFinderRepo workbookFinder,
	authChecker authorizationChecker,
	activeListFinder activeQuestionListFinder,
	activeListSaver activeQuestionListSaver,
	parser contentParser,
) *Command {
	return &Command{
		AddQuestionCommand:    NewAddQuestionCommand(questionSaverRepo, activeListFinder, activeListSaver, authChecker, parser),
		GetQuestionQuery:      NewGetQuestionQuery(questionFinderRepo, workbookFinderRepo, authChecker),
		ListQuestionsQuery:    NewListQuestionsQuery(questionFinderRepo, workbookFinderRepo, authChecker),
		UpdateQuestionCommand: NewUpdateQuestionCommand(questionFinderRepo, questionSaverRepo, authChecker, parser),
		DeleteQuestionCommand: NewDeleteQuestionCommand(questionDeleterRepo, activeListFinder, activeListSaver, authChecker),
	}
}
