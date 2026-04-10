package user

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
)

type appUserFinder interface {
	FindByID(ctx context.Context, id domain.AppUserID) (*domainuser.AppUser, error)
}

type appUserSaver interface {
	Save(ctx context.Context, user *domainuser.AppUser) error
}

// ChangePasswordCommand changes a user's password.
type ChangePasswordCommand struct {
	appUserFinder appUserFinder
	appUserSaver  appUserSaver
	hasher        passwordHasher
	authChecker   authorizationChecker
}

// NewChangePasswordCommand returns a new ChangePasswordCommand.
func NewChangePasswordCommand(
	finder appUserFinder,
	saver appUserSaver,
	hasher passwordHasher,
	authChecker authorizationChecker,
) *ChangePasswordCommand {
	return &ChangePasswordCommand{
		appUserFinder: finder,
		appUserSaver:  saver,
		hasher:        hasher,
		authChecker:   authChecker,
	}
}

// ChangePassword changes the password for the specified user.
func (c *ChangePasswordCommand) ChangePassword(ctx context.Context, input *userservice.ChangePasswordInput) (*userservice.ChangePasswordOutput, error) {
	user, err := c.appUserFinder.FindByID(ctx, input.AppUserID)
	if err != nil {
		return nil, fmt.Errorf("find app user: %w", err)
	}

	// Authorization check using the user's organization.
	allowed, err := c.authChecker.IsAllowed(ctx, user.OrganizationID(), input.OperatorID, domainrbac.ActionChangePassword(), domainrbac.ResourceUser(input.AppUserID))
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
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
