package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
)

// findEnabledAppUser returns domain.ErrUnauthenticated for a missing organization,
// a missing user, or a disabled user so that callers cannot tell them apart.
func findEnabledAppUser(ctx context.Context, organizationFinder OrganizationFinder, appUserFinder AppUserByLoginIDFinder, organizationName string, loginID domain.LoginID) (*domain.Organization, *domainuser.AppUser, error) {
	org, err := organizationFinder.FindByName(ctx, organizationName)
	if errors.Is(err, domain.ErrOrganizationNotFound) {
		return nil, nil, domain.ErrUnauthenticated
	}
	if err != nil {
		return nil, nil, fmt.Errorf("find organization: %w", err)
	}

	user, err := appUserFinder.FindByLoginID(ctx, org.ID(), loginID)
	if errors.Is(err, domain.ErrAppUserNotFound) {
		return nil, nil, domain.ErrUnauthenticated
	}
	if err != nil {
		return nil, nil, fmt.Errorf("find app user in organization %s: %w", org.ID(), err)
	}

	if !user.Enabled() {
		return nil, nil, domain.ErrUnauthenticated
	}
	return org, user, nil
}
