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
) *Command {
	return &Command{
		AddQuestionCommand:    NewAddQuestionCommand(questionAdderRepo, workbookFinderRepo, authChecker),
		GetQuestionQuery:      NewGetQuestionQuery(questionFinderRepo, workbookFinderRepo, authChecker),
		ListQuestionsQuery:    NewListQuestionsQuery(questionFinderRepo, workbookFinderRepo, authChecker),
		UpdateQuestionCommand: NewUpdateQuestionCommand(questionFinderRepo, questionUpdaterRepo, workbookFinderRepo, authChecker),
		DeleteQuestionCommand: NewDeleteQuestionCommand(questionDeleterRepo, workbookFinderRepo, authChecker),
	}
}
