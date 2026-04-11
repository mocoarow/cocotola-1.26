package space

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
)

type spaceSaver interface {
	Save(ctx context.Context, space *domainspace.Space) error
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

// CreateSpaceCommand creates a new space within an organization.
type CreateSpaceCommand struct {
	spaceRepo   spaceSaver
	orgRepo     organizationFinderByName
	publisher   eventPublisher
	authChecker authorizationChecker
}

// NewCreateSpaceCommand returns a new CreateSpaceCommand.
func NewCreateSpaceCommand(
	spaceRepo spaceSaver,
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

	allowed, err := c.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionCreateSpace(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	st, err := domainspace.NewType(input.SpaceType)
	if err != nil {
		return nil, fmt.Errorf("new space type: %w", err)
	}

	var keyName string
	if st.IsPublic() {
		keyName = domainspace.PublicSpaceKeyName(input.OrganizationName)
	} else {
		return nil, errors.New("private spaces must be created via event handler")
	}

	s, err := domainspace.Provision(ctx, c.spaceRepo, org.ID(), input.OperatorID, keyName, input.Name, st)
	if err != nil {
		return nil, fmt.Errorf("provision space: %w", err)
	}

	c.publisher.Publish(domain.NewSpaceCreated(s.ID(), org.ID(), input.OperatorID, keyName, input.Name, input.SpaceType, time.Now()))

	output, err := spaceservice.NewCreateSpaceOutput(s.ID(), org.ID(), input.OperatorID, keyName, input.Name, input.SpaceType, false)
	if err != nil {
		return nil, fmt.Errorf("create space output: %w", err)
	}
	return output, nil
}
