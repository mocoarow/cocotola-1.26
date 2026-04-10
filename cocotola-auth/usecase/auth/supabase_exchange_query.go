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
	appUserFinder          AppUserProviderFinder
	appUserByLoginIDFinder AppUserByLoginIDFinder
	appUserSaver           AppUserSaver
	organizationFinder     OrganizationFinder
}

// NewSupabaseExchangeQuery returns a new SupabaseExchangeQuery.
func NewSupabaseExchangeQuery(
	supabaseVerifier SupabaseVerifier,
	appUserFinder AppUserProviderFinder,
	appUserByLoginIDFinder AppUserByLoginIDFinder,
	appUserSaver AppUserSaver,
	organizationFinder OrganizationFinder,
) *SupabaseExchangeQuery {
	return &SupabaseExchangeQuery{
		supabaseVerifier:       supabaseVerifier,
		appUserFinder:          appUserFinder,
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

	appUser, err := q.appUserFinder.FindByProviderID(ctx, org.ID(), providerName, sub)
	if err == nil {
		return q.buildOutput(appUser.ID(), string(appUser.LoginID()), input.OrganizationName)
	}

	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return nil, fmt.Errorf("find app user by provider id: %w", err)
	}

	return q.findOrCreateUser(ctx, org.ID(), email, providerName, sub, input.OrganizationName)
}

func (q *SupabaseExchangeQuery) findOrCreateUser(ctx context.Context, orgID domain.OrganizationID, email string, provider string, providerID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	loginID := domain.LoginID(email)

	newUser, createErr := domainuser.Provision(ctx, q.appUserSaver, orgID, loginID, "", provider, providerID, true)
	if createErr == nil {
		return q.buildOutput(newUser.ID(), string(newUser.LoginID()), orgName)
	}

	// Race-condition: another request may have created the same provider mapping concurrently.
	if appUser, retryErr := q.appUserFinder.FindByProviderID(ctx, orgID, provider, providerID); retryErr == nil {
		return q.buildOutput(appUser.ID(), string(appUser.LoginID()), orgName)
	}

	// C2: only attempt to link an existing local account when the create actually failed
	// because of a duplicate login_id. Any other error (network, CAS failure, driver fault)
	// must propagate so we do not silently paper over persistence bugs.
	if !errors.Is(createErr, gorm.ErrDuplicatedKey) {
		return nil, fmt.Errorf("save new app user with provider: %w", createErr)
	}

	// Account linking: a user may already exist with the same login ID (e.g. password signup).
	// Load the aggregate, let the domain enforce the linking invariant, then save it.
	existing, findErr := q.appUserByLoginIDFinder.FindByLoginID(ctx, orgID, loginID)
	if findErr != nil {
		if errors.Is(findErr, domain.ErrAppUserNotFound) {
			return nil, fmt.Errorf("save new app user with provider: %w", createErr)
		}
		return nil, fmt.Errorf("find app user by login id for linking: %w", findErr)
	}

	// C1: refuse to auto-link an existing password-holding account.
	if existing.HashedPassword() != "" {
		return nil, fmt.Errorf("app user %s login=%s: %w", existing.ID().String(), email, domain.ErrAppUserAutoLinkRejected)
	}

	if err := existing.LinkProvider(provider, providerID); err != nil {
		return nil, fmt.Errorf("link provider to app user %s: %w", existing.ID().String(), err)
	}

	if err := q.appUserSaver.Save(ctx, existing); err != nil {
		return nil, fmt.Errorf("save linked app user %s: %w", existing.ID().String(), err)
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
