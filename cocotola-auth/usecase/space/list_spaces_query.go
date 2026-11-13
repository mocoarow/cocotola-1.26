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
	FindByID(ctx context.Context, id domain.SpaceID) (*domainspace.Space, error)
	FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) ([]domainspace.Space, error)
	FindPublicOrOwnedByOrganizationID(ctx context.Context, organizationID domain.OrganizationID, ownerID domain.AppUserID) ([]domainspace.Space, error)
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

	allowed, err := q.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionListSpaces(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	canViewAnySpace, err := q.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionViewSpace(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("check view any space: %w", err)
	}
	if canViewAnySpace {
		allSpaces, err := q.spaceRepo.FindByOrganizationID(ctx, org.ID())
		if err != nil {
			return nil, fmt.Errorf("find spaces by organization: %w", err)
		}
		return &spaceservice.ListSpacesOutput{Spaces: toItems(allSpaces)}, nil
	}

	candidates, err := q.spaceRepo.FindPublicOrOwnedByOrganizationID(ctx, org.ID(), input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("find public or owned spaces: %w", err)
	}

	visible := make([]domainspace.Space, 0, len(candidates))
	for _, s := range candidates {
		if !s.SpaceType().IsPublic() {
			spaceAllowed, err := q.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(s.ID()))
			if err != nil {
				return nil, fmt.Errorf("check space access: %w", err)
			}
			if !spaceAllowed {
				continue
			}
		}
		visible = append(visible, s)
	}

	return &spaceservice.ListSpacesOutput{Spaces: toItems(visible)}, nil
}

func toItems(spaces []domainspace.Space) []spaceservice.Item {
	items := make([]spaceservice.Item, len(spaces))
	for i := range spaces {
		s := &spaces[i]
		items[i] = spaceservice.Item{
			SpaceID:        s.ID(),
			OrganizationID: s.OrganizationID(),
			OwnerID:        s.OwnerID(),
			KeyName:        s.KeyName(),
			Name:           s.Name(),
			SpaceType:      s.SpaceType().Value(),
			Deleted:        s.Deleted(),
		}
	}
	return items
}
