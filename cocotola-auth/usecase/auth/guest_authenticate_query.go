package auth

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// GuestAuthenticateQuery verifies guest credentials.
type GuestAuthenticateQuery struct {
	organizationFinder OrganizationFinder
	appUserFinder      AppUserByLoginIDFinder
}

// NewGuestAuthenticateQuery returns a new GuestAuthenticateQuery.
func NewGuestAuthenticateQuery(organizationFinder OrganizationFinder, appUserFinder AppUserByLoginIDFinder) *GuestAuthenticateQuery {
	return &GuestAuthenticateQuery{
		organizationFinder: organizationFinder,
		appUserFinder:      appUserFinder,
	}
}

// GuestAuthenticate verifies guest credentials and returns user info.
func (q *GuestAuthenticateQuery) GuestAuthenticate(ctx context.Context, input *authservice.GuestAuthenticateInput) (*authservice.GuestAuthenticateOutput, error) {
	loginID := domain.LoginID(domainuser.NewGuestLoginID(input.OrganizationName))
	org, user, err := findEnabledAppUser(ctx, q.organizationFinder, q.appUserFinder, input.OrganizationName, loginID)
	if err != nil {
		return nil, fmt.Errorf("authenticate guest: %w", err)
	}

	output, err := authservice.NewGuestAuthenticateOutput(user.ID(), string(user.LoginID()), org.Name())
	if err != nil {
		return nil, fmt.Errorf("create guest authenticate output: %w", err)
	}
	return output, nil
}
