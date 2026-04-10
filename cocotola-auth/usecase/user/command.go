package user

// Command composes all user management Command structs.
type Command struct {
	*CreateAppUserCommand
	*ChangePasswordCommand
}

// NewCommand returns a new Command with the given dependencies.
func NewCommand(
	saver appUserSaver,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	appUserFinder appUserFinder,
	hasher passwordHasher,
	authChecker authorizationChecker,
) *Command {
	return &Command{
		CreateAppUserCommand:  NewCreateAppUserCommand(saver, orgRepo, publisher, hasher, authChecker),
		ChangePasswordCommand: NewChangePasswordCommand(appUserFinder, saver, hasher, authChecker),
	}
}
