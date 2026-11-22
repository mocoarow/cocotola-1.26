package auth

import (
	"context"
	"fmt"
	"slices"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// PasswordAuthenticateQuery verifies user credentials.
type PasswordAuthenticateQuery struct {
	organizationFinder OrganizationFinder
	appUserFinder      AppUserByLoginIDFinder
	groupFinder        GroupFinder
	hasher             PasswordHasher
}

// NewPasswordAuthenticateQuery returns a new PasswordAuthenticateQuery.
func NewPasswordAuthenticateQuery(organizationFinder OrganizationFinder, appUserFinder AppUserByLoginIDFinder, groupFinder GroupFinder, hasher PasswordHasher) *PasswordAuthenticateQuery {
	return &PasswordAuthenticateQuery{
		organizationFinder: organizationFinder,
		appUserFinder:      appUserFinder,
		groupFinder:        groupFinder,
		hasher:             hasher,
	}
}

// PasswordAuthenticate verifies user credentials and returns user info.
func (q *PasswordAuthenticateQuery) PasswordAuthenticate(ctx context.Context, input *authservice.PasswordAuthenticateInput) (*authservice.PasswordAuthenticateOutput, error) {
	org, user, err := findEnabledAppUser(ctx, q.organizationFinder, q.appUserFinder, input.OrganizationName, domain.LoginID(input.LoginID))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	if err := user.VerifyPassword(input.Password, q.hasher); err != nil {
		return nil, domain.ErrUnauthenticated
	}

	groups, err := q.groupFinder.GetGroupsForUser(ctx, org.ID(), user.ID())
	if err != nil {
		return nil, fmt.Errorf("get groups for user %s: %w", user.ID(), err)
	}
	if slices.ContainsFunc(groups, domainrbac.IsLoginDenied) {
		return nil, domain.ErrUnauthenticated
	}

	output, err := authservice.NewPasswordAuthenticateOutput(user.ID(), string(user.LoginID()), org.Name())
	if err != nil {
		return nil, fmt.Errorf("create authenticate output: %w", err)
	}
	return output, nil
}
