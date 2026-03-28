package space

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
)

type spaceFinder interface {
	FindByID(ctx context.Context, id int) (*domainspace.Space, error)
	FindByOrganizationID(ctx context.Context, organizationID int) ([]domainspace.Space, error)
}

// ListSpacesQuery returns spaces accessible by the operator.
type ListSpacesQuery struct {
	spaceRepo   spaceFinder
	orgRepo     organizationFinderByName
	authChecker authorizationChecker
}

// NewListSpacesQuery returns a new ListSpacesQuery.
func NewListSpacesQuery(
	spaceRepo spaceFinder,
	orgRepo organizationFinderByName,
	authChecker authorizationChecker,
) *ListSpacesQuery {
	return &ListSpacesQuery{
		spaceRepo:   spaceRepo,
		orgRepo:     orgRepo,
		authChecker: authChecker,
	}
}

// ListSpaces returns spaces accessible by the operator.
func (q *ListSpacesQuery) ListSpaces(ctx context.Context, input *spaceservice.ListSpacesInput) (*spaceservice.ListSpacesOutput, error) {
	org, err := q.orgRepo.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	allowed, err := q.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionViewSpace(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	allSpaces, err := q.spaceRepo.FindByOrganizationID(ctx, org.ID())
	if err != nil {
		return nil, fmt.Errorf("find spaces by organization: %w", err)
	}

	var items []spaceservice.Item
	for _, s := range allSpaces {
		if s.SpaceType().IsPublic() {
			items = append(items, spaceservice.Item{
				SpaceID:        s.ID(),
				OrganizationID: s.OrganizationID(),
				OwnerID:        s.OwnerID(),
				KeyName:        s.KeyName(),
				Name:           s.Name(),
				SpaceType:      s.SpaceType().Value(),
				Deleted:        s.Deleted(),
			})

			continue
		}

		spaceAllowed, err := q.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(s.ID()))
		if err != nil {
			return nil, fmt.Errorf("check space access: %w", err)
		}
		if spaceAllowed {
			items = append(items, spaceservice.Item{
				SpaceID:        s.ID(),
				OrganizationID: s.OrganizationID(),
				OwnerID:        s.OwnerID(),
				KeyName:        s.KeyName(),
				Name:           s.Name(),
				SpaceType:      s.SpaceType().Value(),
				Deleted:        s.Deleted(),
			})
		}
	}

	return &spaceservice.ListSpacesOutput{Spaces: items}, nil
}
