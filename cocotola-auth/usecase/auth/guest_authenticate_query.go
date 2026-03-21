package auth

import (
	"context"
	"fmt"

	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// GuestAuthenticator verifies guest credentials and returns user info.
type GuestAuthenticator interface {
	Authenticate(ctx context.Context, organizationName string) (*authservice.UserInfo, error)
}

// GuestAuthenticateQuery verifies guest credentials.
type GuestAuthenticateQuery struct {
	guestAuthenticator GuestAuthenticator
}

// NewGuestAuthenticateQuery returns a new GuestAuthenticateQuery.
func NewGuestAuthenticateQuery(guestAuthenticator GuestAuthenticator) *GuestAuthenticateQuery {
	return &GuestAuthenticateQuery{
		guestAuthenticator: guestAuthenticator,
	}
}

// GuestAuthenticate verifies guest credentials and returns user info.
func (q *GuestAuthenticateQuery) GuestAuthenticate(ctx context.Context, input *authservice.GuestAuthenticateInput) (*authservice.GuestAuthenticateOutput, error) {
	userInfo, err := q.guestAuthenticator.Authenticate(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	output, err := authservice.NewGuestAuthenticateOutput(userInfo.UserID, userInfo.LoginID, userInfo.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("create guest authenticate output: %w", err)
	}
	return output, nil
}
