package group

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// Saver persists a Group aggregate as a whole.
type Saver interface {
	Save(ctx context.Context, group *Group) error
}

// Provision generates a fresh UUIDv7 ID, constructs a Group via the domain
// factory (which enforces invariants), and persists it via Saver.
func Provision(
	ctx context.Context,
	saver Saver,
	organizationID domain.OrganizationID,
	name string,
) (*Group, error) {
	id, err := domain.NewGroupIDV7()
	if err != nil {
		return nil, fmt.Errorf("generate group id: %w", err)
	}
	group, err := NewGroup(id, organizationID, name, true)
	if err != nil {
		return nil, fmt.Errorf("new group: %w", err)
	}
	if err := saver.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("save group: %w", err)
	}
	return group, nil
}
