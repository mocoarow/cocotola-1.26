package group

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	groupservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/group"
)

type groupSaver interface {
	Save(ctx context.Context, group *domaingroup.Group) error
}

type organizationFinderByName interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
}

type eventPublisher interface {
	Publish(event domain.Event)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID domain.OrganizationID, operatorID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}

// CreateGroupCommand creates a new group within an organization.
type CreateGroupCommand struct {
	groupRepo   groupSaver
	orgRepo     organizationFinderByName
	publisher   eventPublisher
	authChecker authorizationChecker
}

// NewCreateGroupCommand returns a new CreateGroupCommand.
func NewCreateGroupCommand(
	groupRepo groupSaver,
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
	org, err := c.orgRepo.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	allowed, err := c.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionCreateGroup(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	group, err := domaingroup.Provision(ctx, c.groupRepo, org.ID(), input.GroupName)
	if err != nil {
		return nil, fmt.Errorf("provision group: %w", err)
	}

	c.publisher.Publish(domain.NewGroupCreated(group.ID(), org.ID(), input.GroupName, time.Now()))

	output, err := groupservice.NewCreateGroupOutput(group.ID(), org.ID(), input.GroupName, true)
	if err != nil {
		return nil, fmt.Errorf("create group output: %w", err)
	}
	return output, nil
}
