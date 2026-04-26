// Package space provides use case implementations for space management.
package space

// Command composes all space management Command and Query structs.
type Command struct {
	*CreateSpaceCommand
	*ListSpacesQuery
	*FindSpaceQuery
}

// NewCommand returns a new Command with the given dependencies.
func NewCommand(
	spaceRepo spaceSaver,
	spaceFinderRepo spaceFinder,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	authChecker authorizationChecker,
) *Command {
	return &Command{
		CreateSpaceCommand: NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker),
		ListSpacesQuery:    NewListSpacesQuery(spaceFinderRepo, orgRepo, authChecker),
		FindSpaceQuery:     NewFindSpaceQuery(spaceFinderRepo),
	}
}
