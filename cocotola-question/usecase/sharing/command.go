// Package sharing provides use cases for workbook sharing and reference management.
package sharing

// Command composes all sharing management use cases.
type Command struct {
	*ShareWorkbookCommand
	*ListSharedQuery
	*UnshareCommand
	*ListPublicQuery
}

// NewCommand returns a new Command composing all sharing use cases.
func NewCommand(
	referenceSaverRepo referenceSaver,
	referenceFinderRepo referenceFinder,
	referenceDeleterRepo referenceDeleter,
	workbookFinderRepo workbookFinder,
	publicWorkbookFinderRepo publicWorkbookFinder,
	authChecker authorizationChecker,
) *Command {
	return &Command{
		ShareWorkbookCommand: NewShareWorkbookCommand(referenceSaverRepo, workbookFinderRepo, authChecker),
		ListSharedQuery:      NewListSharedQuery(referenceFinderRepo),
		UnshareCommand:       NewUnshareCommand(referenceDeleterRepo),
		ListPublicQuery:      NewListPublicQuery(publicWorkbookFinderRepo),
	}
}
