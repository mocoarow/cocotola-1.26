package user

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
)

type appUserCreator interface {
	Create(ctx context.Context, organizationID int, loginID string, hashedPassword string) (int, error)
}

type organizationFinder interface {
	FindByID(ctx context.Context, id int) (*domain.Organization, error)
}

type activeUserListRepository interface {
	FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveUserList, error)
	Save(ctx context.Context, list *domain.ActiveUserList) error
}

type passwordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword string, password string) error
}

// CreateAppUserCommand creates a new app user within an organization.
type CreateAppUserCommand struct {
	appUserRepo    appUserCreator
	orgRepo        organizationFinder
	activeUserRepo activeUserListRepository
	hasher         passwordHasher
}

// NewCreateAppUserCommand returns a new CreateAppUserCommand.
func NewCreateAppUserCommand(
	appUserRepo appUserCreator,
	orgRepo organizationFinder,
	activeUserRepo activeUserListRepository,
	hasher passwordHasher,
) *CreateAppUserCommand {
	return &CreateAppUserCommand{
		appUserRepo:    appUserRepo,
		orgRepo:        orgRepo,
		activeUserRepo: activeUserRepo,
		hasher:         hasher,
	}
}

// CreateAppUser creates a new app user, enforcing the organization's active user limit.
func (c *CreateAppUserCommand) CreateAppUser(ctx context.Context, input *userservice.CreateAppUserInput) (*userservice.CreateAppUserOutput, error) {
	// TX1: Find organization to validate existence and get maxActiveUsers.
	org, err := c.orgRepo.FindByID(ctx, input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	// Hash password via domain policy (enforces MinPasswordLength).
	hashedPassword, err := domain.HashPassword(input.Password, c.hasher)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// TX2: Create app user record.
	appUserID, err := c.appUserRepo.Create(ctx, input.OrganizationID, input.LoginID, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("create app user: %w", err)
	}

	// TX3: Update active user list (separate aggregate, separate transaction).
	activeUserList, err := c.activeUserRepo.FindByOrganizationID(ctx, input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("find active user list: %w", err)
	}

	if err := activeUserList.Add(appUserID, org.MaxActiveUsers()); err != nil {
		return nil, fmt.Errorf("add to active user list: %w", err)
	}

	if err := c.activeUserRepo.Save(ctx, activeUserList); err != nil {
		return nil, fmt.Errorf("save active user list: %w", err)
	}

	output, err := userservice.NewCreateAppUserOutput(appUserID)
	if err != nil {
		return nil, fmt.Errorf("create app user output: %w", err)
	}
	return output, nil
}
