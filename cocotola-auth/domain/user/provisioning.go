package user

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// Saver persists an AppUser aggregate as a whole.
type Saver interface {
	Save(ctx context.Context, user *AppUser) error
}

// Provision creates and persists a new AppUser.
// It is the single provisioning path so no caller can bypass aggregate invariants.
func Provision(
	ctx context.Context,
	saver Saver,
	organizationID domain.OrganizationID,
	loginID domain.LoginID,
	hashedPassword string,
	enabled bool,
) (*AppUser, error) {
	id, err := domain.NewAppUserIDV7()
	if err != nil {
		return nil, fmt.Errorf("generate app user id: %w", err)
	}
	user, err := NewAppUser(id, organizationID, loginID, hashedPassword, enabled)
	if err != nil {
		return nil, fmt.Errorf("new app user: %w", err)
	}
	if err := saver.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save app user: %w", err)
	}
	return user, nil
}
