package user

// Command composes all user management Command structs.
type Command struct {
	*CreateAppUserCommand
	*ChangePasswordCommand
}

// NewCommand returns a new Command with the given dependencies.
func NewCommand(
	appUserRepo appUserCreator,
	orgRepo organizationFinder,
	activeUserRepo activeUserListRepository,
	appUserFinder appUserFinder,
	appUserSaver appUserSaver,
	hasher passwordHasher,
) *Command {
	return &Command{
		CreateAppUserCommand:  NewCreateAppUserCommand(appUserRepo, orgRepo, activeUserRepo, hasher),
		ChangePasswordCommand: NewChangePasswordCommand(appUserFinder, appUserSaver, hasher),
	}
}
