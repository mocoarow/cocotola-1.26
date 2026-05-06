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
	workbookSaverRepo workbookSaver,
	workbookFinderRepo workbookFinder,
	workbookDeleterRepo workbookDeleter,
	ownedListFinder ownedWorkbookListFinder,
	ownedListSaver ownedWorkbookListSaver,
	maxWbFetcher maxWorkbooksFetcher,
	spaceTypeFetcher spaceTypeFetcher,
	authChecker authorizationChecker,
	policyAdder policyAdder,
) *Command {
	return &Command{
		CreateWorkbookCommand: NewCreateWorkbookCommand(workbookSaverRepo, ownedListFinder, ownedListSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder),
		GetWorkbookQuery:      NewGetWorkbookQuery(workbookFinderRepo, authChecker),
		ListWorkbooksQuery:    NewListWorkbooksQuery(workbookFinderRepo, authChecker),
		UpdateWorkbookCommand: NewUpdateWorkbookCommand(workbookFinderRepo, workbookSaverRepo, authChecker),
		DeleteWorkbookCommand: NewDeleteWorkbookCommand(workbookFinderRepo, workbookDeleterRepo, ownedListFinder, ownedListSaver, authChecker),
	}
}
