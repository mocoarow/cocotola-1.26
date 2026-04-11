package auth

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// SupabaseExchangeQuery handles Supabase JWT exchange for internal user lookup or creation.
type SupabaseExchangeQuery struct {
	supabaseVerifier       SupabaseVerifier
	providerFinder         AppUserProviderFinder
	providerSaver          AppUserProviderSaver
	appUserByIDFinder      AppUserByIDFinder
	appUserByLoginIDFinder AppUserByLoginIDFinder
	appUserSaver           AppUserSaver
	organizationFinder     OrganizationFinder
}

// NewSupabaseExchangeQuery returns a new SupabaseExchangeQuery.
func NewSupabaseExchangeQuery(
	supabaseVerifier SupabaseVerifier,
	providerFinder AppUserProviderFinder,
	providerSaver AppUserProviderSaver,
	appUserByIDFinder AppUserByIDFinder,
	appUserByLoginIDFinder AppUserByLoginIDFinder,
	appUserSaver AppUserSaver,
	organizationFinder OrganizationFinder,
) *SupabaseExchangeQuery {
	return &SupabaseExchangeQuery{
		supabaseVerifier:       supabaseVerifier,
		providerFinder:         providerFinder,
		providerSaver:          providerSaver,
		appUserByIDFinder:      appUserByIDFinder,
		appUserByLoginIDFinder: appUserByLoginIDFinder,
		appUserSaver:           appUserSaver,
		organizationFinder:     organizationFinder,
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

	// Step 1: Check if a provider link already exists.
	providerLink, err := q.providerFinder.FindByProviderID(ctx, org.ID(), providerName, sub)
	if err == nil {
		// Provider link found — load the app user and return.
		appUser, findErr := q.appUserByIDFinder.FindByID(ctx, providerLink.AppUserID())
		if findErr != nil {
			return nil, fmt.Errorf("find app user by id: %w", findErr)
		}
		return q.buildOutput(appUser.ID(), string(appUser.LoginID()), input.OrganizationName)
	}

	if !errors.Is(err, domain.ErrAppUserProviderNotFound) {
		return nil, fmt.Errorf("find provider link: %w", err)
	}

	// Step 2: No provider link. Find or create user, then create the provider link.
	return q.findOrCreateUserAndLink(ctx, org.ID(), email, providerName, sub, input.OrganizationName)
}

func (q *SupabaseExchangeQuery) findOrCreateUserAndLink(ctx context.Context, orgID domain.OrganizationID, email string, provider string, providerID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	loginID := domain.LoginID(email)

	// Try to create a new user (without provider — provider is now separate).
	newUser, createErr := domainuser.Provision(ctx, q.appUserSaver, orgID, loginID, "", true)
	if createErr == nil {
		// User created — now create the provider link.
		_, linkErr := domainuser.ProvisionProvider(ctx, q.providerSaver, newUser.ID(), orgID, provider, providerID)
		if linkErr != nil {
			return nil, fmt.Errorf("create provider link for new user: %w", linkErr)
		}
		return q.buildOutput(newUser.ID(), string(newUser.LoginID()), orgName)
	}

	// Race-condition: another request may have created the same provider mapping concurrently.
	if providerLink, retryErr := q.providerFinder.FindByProviderID(ctx, orgID, provider, providerID); retryErr == nil {
		appUser, findErr := q.appUserByIDFinder.FindByID(ctx, providerLink.AppUserID())
		if findErr != nil {
			return nil, fmt.Errorf("find app user by id after retry: %w", findErr)
		}
		return q.buildOutput(appUser.ID(), string(appUser.LoginID()), orgName)
	}

	// Only attempt to link an existing local account when the create failed
	// because of a duplicate login_id.
	if !errors.Is(createErr, gorm.ErrDuplicatedKey) {
		return nil, fmt.Errorf("save new app user: %w", createErr)
	}

	// Account linking: a user may already exist with the same login ID (e.g. password signup).
	existing, findErr := q.appUserByLoginIDFinder.FindByLoginID(ctx, orgID, loginID)
	if findErr != nil {
		if errors.Is(findErr, domain.ErrAppUserNotFound) {
			return nil, fmt.Errorf("save new app user: %w", createErr)
		}
		return nil, fmt.Errorf("find app user by login id for linking: %w", findErr)
	}

	// Refuse to auto-link an existing password-holding account.
	if existing.HashedPassword() != "" {
		return nil, fmt.Errorf("app user %s login=%s: %w", existing.ID().String(), email, domain.ErrAppUserAutoLinkRejected)
	}

	// Create the provider link for the existing user.
	_, linkErr := domainuser.ProvisionProvider(ctx, q.providerSaver, existing.ID(), orgID, provider, providerID)
	if linkErr != nil {
		return nil, fmt.Errorf("link provider to app user %s: %w", existing.ID().String(), linkErr)
	}

	return q.buildOutput(existing.ID(), string(existing.LoginID()), orgName)
}

func (q *SupabaseExchangeQuery) buildOutput(userID domain.AppUserID, loginID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	output, err := authservice.NewSupabaseExchangeOutput(userID, loginID, orgName)
	if err != nil {
		return nil, fmt.Errorf("new supabase exchange output: %w", err)
	}

	return output, nil
}
