// Package workbook provides use cases for workbook management.
package workbook

// Command composes all workbook management use cases.
type Command struct {
	*CreateWorkbookCommand
	*GetWorkbookQuery
	*ListWorkbooksQuery
	*UpdateWorkbookCommand
	*DeleteWorkbookCommand
}

// NewCommand returns a new Command composing all workbook use cases.
func NewCommand(
	workbookCreatorRepo workbookCreator,
	workbookFinderRepo workbookFinder,
	workbookUpdaterRepo workbookUpdater,
	workbookDeleterRepo workbookDeleter,
	authChecker authorizationChecker,
) *Command {
	return &Command{
		CreateWorkbookCommand: NewCreateWorkbookCommand(workbookCreatorRepo, authChecker),
		GetWorkbookQuery:      NewGetWorkbookQuery(workbookFinderRepo, authChecker),
		ListWorkbooksQuery:    NewListWorkbooksQuery(workbookFinderRepo, authChecker),
		UpdateWorkbookCommand: NewUpdateWorkbookCommand(workbookFinderRepo, workbookUpdaterRepo, authChecker),
		DeleteWorkbookCommand: NewDeleteWorkbookCommand(workbookFinderRepo, workbookDeleterRepo, authChecker),
	}
}
