package space

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
)

type spaceByIDFinder interface {
	FindByID(ctx context.Context, id domain.SpaceID) (*domainspace.Space, error)
}

// FindSpaceQuery resolves a single space record by ID for internal callers.
// Authorization is enforced by the X-Service-Api-Key middleware, so no
// per-user RBAC check is performed here.
type FindSpaceQuery struct {
	spaceRepo spaceByIDFinder
}

// NewFindSpaceQuery returns a new FindSpaceQuery.
func NewFindSpaceQuery(spaceRepo spaceByIDFinder) *FindSpaceQuery {
	return &FindSpaceQuery{
		spaceRepo: spaceRepo,
	}
}

// FindSpace returns the space identified by the given ID.
func (q *FindSpaceQuery) FindSpace(ctx context.Context, input *spaceservice.FindSpaceInput) (*spaceservice.FindSpaceOutput, error) {
	s, err := q.spaceRepo.FindByID(ctx, input.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("find space by id: %w", err)
	}

	return &spaceservice.FindSpaceOutput{
		Item: spaceservice.Item{
			SpaceID:        s.ID(),
			OrganizationID: s.OrganizationID(),
			OwnerID:        s.OwnerID(),
			KeyName:        s.KeyName(),
			Name:           s.Name(),
			SpaceType:      s.SpaceType().Value(),
			Deleted:        s.Deleted(),
		},
	}, nil
}
