package user

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
)

type appUserCreator interface {
	Create(ctx context.Context, organizationID int, loginID string, hashedPassword string) (int, error)
}

type organizationFinder interface {
	FindByID(ctx context.Context, id int) (*domain.Organization, error)
}

type eventPublisher interface {
	Publish(event domain.Event)
}

type passwordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword string, password string) error
}

// CreateAppUserCommand creates a new app user within an organization.
type CreateAppUserCommand struct {
	appUserRepo appUserCreator
	orgRepo     organizationFinder
	publisher   eventPublisher
	hasher      passwordHasher
}

// NewCreateAppUserCommand returns a new CreateAppUserCommand.
func NewCreateAppUserCommand(
	appUserRepo appUserCreator,
	orgRepo organizationFinder,
	publisher eventPublisher,
	hasher passwordHasher,
) *CreateAppUserCommand {
	return &CreateAppUserCommand{
		appUserRepo: appUserRepo,
		orgRepo:     orgRepo,
		publisher:   publisher,
		hasher:      hasher,
	}
}

// CreateAppUser creates a new app user and publishes an AppUserCreated event.
func (c *CreateAppUserCommand) CreateAppUser(ctx context.Context, input *userservice.CreateAppUserInput) (*userservice.CreateAppUserOutput, error) {
	// TX1: Find organization to validate existence.
	if _, err := c.orgRepo.FindByID(ctx, input.OrganizationID); err != nil {
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

	// Publish domain event for eventual consistency with ActiveUserList.
	c.publisher.Publish(domain.NewAppUserCreated(appUserID, input.OrganizationID, input.LoginID, time.Now()))

	output, err := userservice.NewCreateAppUserOutput(appUserID)
	if err != nil {
		return nil, fmt.Errorf("create app user output: %w", err)
	}
	return output, nil
}
