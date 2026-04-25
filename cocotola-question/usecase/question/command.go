// Package question provides use cases for question management.
package question

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
	questionAdderRepo questionAdder,
	questionFinderRepo questionFinder,
	questionUpdaterRepo questionUpdater,
	questionDeleterRepo questionDeleter,
	workbookFinderRepo workbookFinder,
	authChecker authorizationChecker,
	activeListFinder activeQuestionListFinder,
	activeListSaver activeQuestionListSaver,
) *Command {
	return &Command{
		AddQuestionCommand:    NewAddQuestionCommand(questionAdderRepo, activeListFinder, activeListSaver, authChecker),
		GetQuestionQuery:      NewGetQuestionQuery(questionFinderRepo, workbookFinderRepo, authChecker),
		ListQuestionsQuery:    NewListQuestionsQuery(questionFinderRepo, workbookFinderRepo, authChecker),
		UpdateQuestionCommand: NewUpdateQuestionCommand(questionFinderRepo, questionUpdaterRepo, authChecker),
		DeleteQuestionCommand: NewDeleteQuestionCommand(questionDeleterRepo, activeListFinder, activeListSaver, authChecker),
	}
}
