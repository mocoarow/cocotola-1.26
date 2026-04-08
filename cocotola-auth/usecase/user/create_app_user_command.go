package user

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
)

type appUserIDProvider interface {
	NextID(ctx context.Context) (int, error)
}

type organizationFinderByName interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
}

type eventPublisher interface {
	Publish(event domain.Event)
}

type passwordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword string, password string) error
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}

// CreateAppUserCommand creates a new app user within an organization.
type CreateAppUserCommand struct {
	idProvider  appUserIDProvider
	saver       appUserSaver
	orgRepo     organizationFinderByName
	publisher   eventPublisher
	hasher      passwordHasher
	authChecker authorizationChecker
}

// NewCreateAppUserCommand returns a new CreateAppUserCommand.
func NewCreateAppUserCommand(
	idProvider appUserIDProvider,
	saver appUserSaver,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	hasher passwordHasher,
	authChecker authorizationChecker,
) *CreateAppUserCommand {
	return &CreateAppUserCommand{
		idProvider:  idProvider,
		saver:       saver,
		orgRepo:     orgRepo,
		publisher:   publisher,
		hasher:      hasher,
		authChecker: authChecker,
	}
}

// CreateAppUser creates a new app user and publishes an AppUserCreated event.
func (c *CreateAppUserCommand) CreateAppUser(ctx context.Context, input *userservice.CreateAppUserInput) (*userservice.CreateAppUserOutput, error) {
	// TX1: Find organization by name to get organizationID.
	org, err := c.orgRepo.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	// Authorization check.
	allowed, err := c.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	// Hash password via domain policy (enforces MinPasswordLength).
	hashedPassword, err := domainuser.HashPassword(input.Password, c.hasher)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// TX2: Reserve aggregate ID, build the aggregate via the domain factory, and persist.
	user, err := domainuser.Provision(ctx, c.idProvider, c.saver, org.ID(), domain.LoginID(input.LoginID), hashedPassword, "", "", true)
	if err != nil {
		return nil, fmt.Errorf("provision app user: %w", err)
	}
	appUserID := user.ID()

	// Publish domain event for eventual consistency with ActiveUserList.
	c.publisher.Publish(domain.NewAppUserCreated(appUserID, org.ID(), input.LoginID, time.Now()))

	output, err := userservice.NewCreateAppUserOutput(appUserID, org.ID(), input.LoginID, true)
	if err != nil {
		return nil, fmt.Errorf("create app user output: %w", err)
	}
	return output, nil
}
