package user

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
)

type appUserFinder interface {
	FindByID(ctx context.Context, id int) (*domain.AppUser, error)
}

type appUserSaver interface {
	Save(ctx context.Context, user *domain.AppUser) error
}

// ChangePasswordCommand changes a user's password.
type ChangePasswordCommand struct {
	appUserFinder appUserFinder
	appUserSaver  appUserSaver
	hasher        passwordHasher
}

// NewChangePasswordCommand returns a new ChangePasswordCommand.
func NewChangePasswordCommand(
	finder appUserFinder,
	saver appUserSaver,
	hasher passwordHasher,
) *ChangePasswordCommand {
	return &ChangePasswordCommand{
		appUserFinder: finder,
		appUserSaver:  saver,
		hasher:        hasher,
	}
}

// ChangePassword changes the password for the specified user.
func (c *ChangePasswordCommand) ChangePassword(ctx context.Context, input *userservice.ChangePasswordInput) (*userservice.ChangePasswordOutput, error) {
	user, err := c.appUserFinder.FindByID(ctx, input.AppUserID)
	if err != nil {
		return nil, fmt.Errorf("find app user: %w", err)
	}

	if err := user.ChangePassword(input.NewPassword, c.hasher); err != nil {
		return nil, fmt.Errorf("change password: %w", err)
	}

	if err := c.appUserSaver.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save app user: %w", err)
	}

	output, err := userservice.NewChangePasswordOutput(input.AppUserID)
	if err != nil {
		return nil, fmt.Errorf("create change password output: %w", err)
	}
	return output, nil
}
