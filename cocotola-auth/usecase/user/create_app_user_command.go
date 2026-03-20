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

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domain.RBACAction, resource domain.RBACResource) (bool, error)
}

// CreateAppUserCommand creates a new app user within an organization.
type CreateAppUserCommand struct {
	appUserRepo appUserCreator
	orgRepo     organizationFinder
	publisher   eventPublisher
	hasher      passwordHasher
	authChecker authorizationChecker
}

// NewCreateAppUserCommand returns a new CreateAppUserCommand.
func NewCreateAppUserCommand(
	appUserRepo appUserCreator,
	orgRepo organizationFinder,
	publisher eventPublisher,
	hasher passwordHasher,
	authChecker authorizationChecker,
) *CreateAppUserCommand {
	return &CreateAppUserCommand{
		appUserRepo: appUserRepo,
		orgRepo:     orgRepo,
		publisher:   publisher,
		hasher:      hasher,
		authChecker: authChecker,
	}
}

// CreateAppUser creates a new app user and publishes an AppUserCreated event.
func (c *CreateAppUserCommand) CreateAppUser(ctx context.Context, input *userservice.CreateAppUserInput) (*userservice.CreateAppUserOutput, error) {
	// Authorization check.
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionCreateUser(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

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

	output, err := userservice.NewCreateAppUserOutput(appUserID, input.OrganizationID, input.LoginID, true)
	if err != nil {
		return nil, fmt.Errorf("create app user output: %w", err)
	}
	return output, nil
}
