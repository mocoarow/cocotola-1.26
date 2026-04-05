package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// SupabaseExchangeQuery handles Supabase JWT exchange for internal user lookup or creation.
type SupabaseExchangeQuery struct {
	supabaseVerifier   SupabaseVerifier
	appUserFinder      AppUserProviderFinder
	appUserCreator     AppUserProviderCreator
	organizationFinder OrganizationFinder
}

// NewSupabaseExchangeQuery returns a new SupabaseExchangeQuery.
func NewSupabaseExchangeQuery(
	supabaseVerifier SupabaseVerifier,
	appUserFinder AppUserProviderFinder,
	appUserCreator AppUserProviderCreator,
	organizationFinder OrganizationFinder,
) *SupabaseExchangeQuery {
	return &SupabaseExchangeQuery{
		supabaseVerifier:   supabaseVerifier,
		appUserFinder:      appUserFinder,
		appUserCreator:     appUserCreator,
		organizationFinder: organizationFinder,
	}
}

// SupabaseExchange verifies a Supabase JWT, finds or creates an internal user, and returns user info.
func (q *SupabaseExchangeQuery) SupabaseExchange(ctx context.Context, input *authservice.SupabaseExchangeInput) (*authservice.SupabaseExchangeOutput, error) {
	sub, email, err := q.supabaseVerifier.Verify(ctx, input.SupabaseJWT)
	if err != nil {
		return nil, fmt.Errorf("verify supabase token: %w", err)
	}

	org, err := q.organizationFinder.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization %s: %w", input.OrganizationName, err)
	}

	const providerName = "supabase"

	appUser, err := q.appUserFinder.FindByProviderID(ctx, org.ID(), providerName, sub)
	if err == nil {
		return q.buildOutput(appUser.ID(), string(appUser.LoginID()), input.OrganizationName)
	}

	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return nil, fmt.Errorf("find app user by provider id: %w", err)
	}

	return q.findOrCreateUser(ctx, org.ID(), email, providerName, sub, input.OrganizationName)
}

func (q *SupabaseExchangeQuery) findOrCreateUser(ctx context.Context, orgID int, email string, provider string, providerID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	userID, err := q.appUserCreator.CreateWithProvider(ctx, orgID, email, provider, providerID)
	if err == nil {
		return q.buildOutput(userID, email, orgName)
	}

	// Handle race condition: another request may have created the user concurrently.
	appUser, retryErr := q.appUserFinder.FindByProviderID(ctx, orgID, provider, providerID)
	if retryErr != nil {
		return nil, fmt.Errorf("create app user with provider: %w", err)
	}

	return q.buildOutput(appUser.ID(), string(appUser.LoginID()), orgName)
}

func (q *SupabaseExchangeQuery) buildOutput(userID int, loginID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	output, err := authservice.NewSupabaseExchangeOutput(userID, loginID, orgName)
	if err != nil {
		return nil, fmt.Errorf("new supabase exchange output: %w", err)
	}

	return output, nil
}
