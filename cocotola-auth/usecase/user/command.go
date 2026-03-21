package user

// Command composes all user management Command structs.
type Command struct {
	*CreateAppUserCommand
	*ChangePasswordCommand
}

// NewCommand returns a new Command with the given dependencies.
func NewCommand(
	appUserRepo appUserCreator,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	appUserFinder appUserFinder,
	appUserSaver appUserSaver,
	hasher passwordHasher,
	authChecker authorizationChecker,
) *Command {
	return &Command{
		CreateAppUserCommand:  NewCreateAppUserCommand(appUserRepo, orgRepo, publisher, hasher, authChecker),
		ChangePasswordCommand: NewChangePasswordCommand(appUserFinder, appUserSaver, hasher, authChecker),
	}
}
