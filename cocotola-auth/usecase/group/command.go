package group

// Command composes all group management Command structs.
type Command struct {
	*CreateGroupCommand
}

// NewCommand returns a new Command with the given dependencies.
func NewCommand(
	groupRepo groupSaver,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	authChecker authorizationChecker,
) *Command {
	return &Command{
		CreateGroupCommand: NewCreateGroupCommand(groupRepo, orgRepo, publisher, authChecker),
	}
}
