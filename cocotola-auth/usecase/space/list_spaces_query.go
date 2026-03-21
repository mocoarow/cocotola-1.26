package space

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
)

type spaceFinder interface {
	FindByID(ctx context.Context, id int) (*domain.Space, error)
	FindByOrganizationID(ctx context.Context, organizationID int) ([]domain.Space, error)
}

type userSpaceFinder interface {
	FindSpaceIDsByUserID(ctx context.Context, organizationID int, userID int) ([]int, error)
}

// ListSpacesQuery returns spaces accessible by the operator.
type ListSpacesQuery struct {
	spaceRepo     spaceFinder
	userSpaceRepo userSpaceFinder
	orgRepo       organizationFinderByName
	authChecker   authorizationChecker
}

// NewListSpacesQuery returns a new ListSpacesQuery.
func NewListSpacesQuery(
	spaceRepo spaceFinder,
	userSpaceRepo userSpaceFinder,
	orgRepo organizationFinderByName,
	authChecker authorizationChecker,
) *ListSpacesQuery {
	return &ListSpacesQuery{
		spaceRepo:     spaceRepo,
		userSpaceRepo: userSpaceRepo,
		orgRepo:       orgRepo,
		authChecker:   authChecker,
	}
}

// ListSpaces returns spaces accessible by the operator.
func (q *ListSpacesQuery) ListSpaces(ctx context.Context, input *spaceservice.ListSpacesInput) (*spaceservice.ListSpacesOutput, error) {
	org, err := q.orgRepo.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization: %w", err)
	}

	allowed, err := q.authChecker.IsAllowed(ctx, org.ID(), input.OperatorID, domain.ActionViewSpace(), domain.ResourceAny())
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

	userSpaceIDs, err := q.userSpaceRepo.FindSpaceIDsByUserID(ctx, org.ID(), input.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("find user space ids: %w", err)
	}

	userSpaceSet := make(map[int]bool, len(userSpaceIDs))
	for _, id := range userSpaceIDs {
		userSpaceSet[id] = true
	}

	var items []spaceservice.Item
	for _, s := range allSpaces {
		if s.SpaceType().IsPublic() || userSpaceSet[s.ID()] {
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
