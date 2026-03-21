package group

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	groupservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/group"
)

type groupCreator interface {
	Create(ctx context.Context, organizationID int, name string) (int, error)
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

// CreateGroupCommand creates a new group within an organization.
type CreateGroupCommand struct {
	groupRepo   groupCreator
	orgRepo     organizationFinderByName
	publisher   eventPublisher
	authChecker authorizationChecker
}

// NewCreateGroupCommand returns a new CreateGroupCommand.
func NewCreateGroupCommand(
	groupRepo groupCreator,
	orgRepo organizationFinderByName,
	publisher eventPublisher,
	authChecker authorizationChecker,
) *CreateGroupCommand {
	return &CreateGroupCommand{
		groupRepo:   groupRepo,
		orgRepo:     orgRepo,
		publisher:   publisher,
		authChecker: authChecker,
	}
}

// CreateGroup creates a new group and publishes a GroupCreated event.
func (c *CreateGroupCommand) CreateGroup(ctx context.Context, input *groupservice.CreateGroupInput) (*groupservice.CreateGroupOutput, error) {
	// TX1: Find organization by name to get organizationID.
	org, err := c.orgRepo.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	// Authorization check.
	allowed, err := c.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domain.ActionCreateGroup(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	// TX2: Create group record.
	groupID, err := c.groupRepo.Create(ctx, org.ID(), input.GroupName)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	// Publish domain event for eventual consistency with ActiveGroupList.
	c.publisher.Publish(domain.NewGroupCreated(groupID, org.ID(), input.GroupName, time.Now()))

	output, err := groupservice.NewCreateGroupOutput(groupID, org.ID(), input.GroupName, true)
	if err != nil {
		return nil, fmt.Errorf("create group output: %w", err)
	}
	return output, nil
}
