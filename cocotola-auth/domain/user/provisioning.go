package user

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// IDProvider reserves a fresh aggregate identifier from the persistence layer.
type IDProvider interface {
	NextID(ctx context.Context) (int, error)
}

// Saver persists an AppUser aggregate as a whole.
type Saver interface {
	Save(ctx context.Context, user *AppUser) error
}

// Provision reserves a new aggregate ID, constructs an AppUser via the domain
// factory (which enforces invariants), and persists it via Saver. This is the
// single provisioning path shared by all callers (create-user command, supabase
// exchange linking, cocotola-init bootstrap) so that no code path can bypass
// aggregate invariants.
func Provision(
	ctx context.Context,
	idProvider IDProvider,
	saver Saver,
	organizationID int,
	loginID domain.LoginID,
	hashedPassword string,
	provider string,
	providerID string,
	enabled bool,
) (*AppUser, error) {
	id, err := idProvider.NextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("next app user id: %w", err)
	}
	user, err := NewAppUser(id, organizationID, loginID, hashedPassword, provider, providerID, enabled)
	if err != nil {
		return nil, fmt.Errorf("new app user: %w", err)
	}
	if err := saver.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save app user: %w", err)
	}
	return user, nil
}
