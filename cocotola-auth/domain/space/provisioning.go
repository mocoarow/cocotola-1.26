package space

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// Saver persists a Space aggregate as a whole.
type Saver interface {
	Save(ctx context.Context, space *Space) error
}

// Provision generates a fresh UUIDv7 ID, constructs a Space via the domain
// factory (which enforces invariants), and persists it via Saver.
func Provision(
	ctx context.Context,
	saver Saver,
	organizationID domain.OrganizationID,
	ownerID domain.AppUserID,
	keyName string,
	name string,
	spaceType Type,
) (*Space, error) {
	id, err := domain.NewSpaceIDV7()
	if err != nil {
		return nil, fmt.Errorf("generate space id: %w", err)
	}
	s, err := NewSpace(id, organizationID, ownerID, keyName, name, spaceType, false)
	if err != nil {
		return nil, fmt.Errorf("new space: %w", err)
	}
	if err := saver.Save(ctx, s); err != nil {
		return nil, fmt.Errorf("save space: %w", err)
	}
	s.IncrementVersion()
	return s, nil
}
