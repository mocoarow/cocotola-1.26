package space

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
)

type spaceCreator interface {
	Create(ctx context.Context, organizationID int, ownerID int, keyName string, name string, spaceType string, createdBy int) (int, error)
}

type organizationFinderByName interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
}

type eventPublisher interface {
	Publish(event domain.Event)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domain.RBACAction, resource domain.RBACResource) (bool, error)
}

// CreateSpaceCommand creates a new space within an organization.
type CreateSpaceCommand struct {
	spaceRepo   spaceCreator
	orgRepo     organizationFinderByName
	publisher   eventPublisher
	authChecker authorizationChecker
}

// NewCreateSpaceCommand returns a new CreateSpaceCommand.
func NewCreateSpaceCommand(
	spaceRepo spaceCreator,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	authChecker authorizationChecker,
) *CreateSpaceCommand {
	return &CreateSpaceCommand{
		spaceRepo:   spaceRepo,
		orgRepo:     orgRepo,
		publisher:   publisher,
		authChecker: authChecker,
	}
}

// CreateSpace creates a new space and publishes a SpaceCreated event.
func (c *CreateSpaceCommand) CreateSpace(ctx context.Context, input *spaceservice.CreateSpaceInput) (*spaceservice.CreateSpaceOutput, error) {
	org, err := c.orgRepo.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	allowed, err := c.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domain.ActionCreateSpace(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	st, err := domain.NewSpaceType(input.SpaceType)
	if err != nil {
		return nil, fmt.Errorf("new space type: %w", err)
	}

	var keyName string
	if st.IsPublic() {
		keyName = domain.PublicSpaceKeyName(input.OrganizationName)
	} else {
		return nil, errors.New("private spaces must be created via event handler")
	}

	spaceID, err := c.spaceRepo.Create(ctx, org.ID(), input.OperatorID, keyName, input.Name, input.SpaceType, input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("create space: %w", err)
	}

	c.publisher.Publish(domain.NewSpaceCreated(spaceID, org.ID(), input.OperatorID, keyName, input.Name, input.SpaceType, time.Now()))

	output, err := spaceservice.NewCreateSpaceOutput(spaceID, org.ID(), input.OperatorID, keyName, input.Name, input.SpaceType, false)
	if err != nil {
		return nil, fmt.Errorf("create space output: %w", err)
	}
	return output, nil
}
