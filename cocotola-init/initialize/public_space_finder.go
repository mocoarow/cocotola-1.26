package initialize

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

// ErrPublicSpaceNotFound is returned when the organization has no public space.
var ErrPublicSpaceNotFound = errors.New("public space not found")

// PublicSpaceFinder looks up an organization's public space without creating it.
type PublicSpaceFinder struct {
	repo *gateway.SpaceRepository
}

// NewPublicSpaceFinder returns a PublicSpaceFinder backed by db.
func NewPublicSpaceFinder(db *gorm.DB) *PublicSpaceFinder {
	return &PublicSpaceFinder{repo: gateway.NewSpaceRepository(db)}
}

// FindPublicSpaceID returns the ID of the organization's public space,
// ignoring deleted spaces.
func (f *PublicSpaceFinder) FindPublicSpaceID(ctx context.Context, organizationID domain.OrganizationID) (domain.SpaceID, error) {
	space, err := f.repo.FindPublicByOrganizationID(ctx, organizationID)
	if errors.Is(err, domain.ErrSpaceNotFound) {
		return domain.SpaceID{}, fmt.Errorf("organization %s: %w", organizationID, ErrPublicSpaceNotFound)
	}
	if err != nil {
		return domain.SpaceID{}, fmt.Errorf("find public space of organization %s: %w", organizationID, err)
	}
	return space.ID(), nil
}
