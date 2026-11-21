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
func (f *PublicSpaceFinder) FindPublicSpaceID(ctx context.Context, organizationID string) (string, error) {
	orgID, err := domain.ParseOrganizationID(organizationID)
	if err != nil {
		return "", fmt.Errorf("parse organization id %q: %w", organizationID, err)
	}
	space, err := f.repo.FindPublicByOrganizationID(ctx, orgID)
	if errors.Is(err, domain.ErrSpaceNotFound) {
		return "", fmt.Errorf("organization %s: %w", orgID, ErrPublicSpaceNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("find public space of organization %s: %w", orgID, err)
	}
	return space.ID().String(), nil
}
