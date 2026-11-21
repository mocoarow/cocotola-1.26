package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// eventPublisher publishes domain events asynchronously.
type eventPublisher interface {
	Publish(event domain.Event)
}

// SupabaseExchangeCommand handles Supabase JWT exchange for internal user lookup or creation.
type SupabaseExchangeCommand struct {
	supabaseVerifier       SupabaseVerifier
	providerFinder         AppUserProviderFinder
	providerSaver          AppUserProviderSaver
	appUserByIDFinder      AppUserByIDFinder
	appUserByLoginIDFinder AppUserByLoginIDFinder
	appUserSaver           AppUserSaver
	organizationFinder     OrganizationFinder
	publisher              eventPublisher
}

// NewSupabaseExchangeCommand returns a new SupabaseExchangeCommand.
func NewSupabaseExchangeCommand(
	supabaseVerifier SupabaseVerifier,
	providerFinder AppUserProviderFinder,
	providerSaver AppUserProviderSaver,
	appUserByIDFinder AppUserByIDFinder,
	appUserByLoginIDFinder AppUserByLoginIDFinder,
	appUserSaver AppUserSaver,
	organizationFinder OrganizationFinder,
	publisher eventPublisher,
) *SupabaseExchangeCommand {
	return &SupabaseExchangeCommand{
		supabaseVerifier:       supabaseVerifier,
		providerFinder:         providerFinder,
		providerSaver:          providerSaver,
		appUserByIDFinder:      appUserByIDFinder,
		appUserByLoginIDFinder: appUserByLoginIDFinder,
		appUserSaver:           appUserSaver,
		organizationFinder:     organizationFinder,
		publisher:              publisher,
	}
}

// SupabaseExchange verifies a Supabase JWT, finds or creates an internal user, and returns user info.
func (c *SupabaseExchangeCommand) SupabaseExchange(ctx context.Context, input *authservice.SupabaseExchangeInput) (*authservice.SupabaseExchangeOutput, error) {
	sub, email, err := c.supabaseVerifier.Verify(ctx, input.SupabaseJWT)
	if err != nil {
		return nil, fmt.Errorf("verify supabase token: %w", err)
	}

	org, err := c.organizationFinder.FindByName(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("find organization %s: %w", input.OrganizationName, err)
	}

	const providerName = "supabase"

	// Step 1: Check if a provider link already exists.
	providerLink, err := c.providerFinder.FindByProviderID(ctx, org.ID(), providerName, sub)
	if err == nil {
		return c.outputForProviderLink(ctx, providerLink, input.OrganizationName)
	}

	if !errors.Is(err, domain.ErrAppUserProviderNotFound) {
		return nil, fmt.Errorf("find provider link: %w", err)
	}

	// Step 2: No provider link. Find or create user, then create the provider link.
	return c.findOrCreateUserAndLink(ctx, org.ID(), email, providerName, sub, input.OrganizationName)
}

func (c *SupabaseExchangeCommand) findOrCreateUserAndLink(ctx context.Context, orgID domain.OrganizationID, email string, provider string, providerID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	loginID := domain.LoginID(email)

	newUser, createErr := domainuser.Provision(ctx, c.appUserSaver, orgID, loginID, "", true)
	if createErr == nil {
		return c.linkNewUser(ctx, newUser, orgID, provider, providerID, orgName)
	}

	// Race-condition: another request may have created the same provider mapping concurrently.
	providerLink, retryErr := c.providerFinder.FindByProviderID(ctx, orgID, provider, providerID)
	if retryErr == nil {
		return c.outputForProviderLink(ctx, providerLink, orgName)
	}
	if !errors.Is(retryErr, domain.ErrAppUserProviderNotFound) {
		return nil, fmt.Errorf("save new app user: %w: find provider link after save: %w: %w", createErr, retryErr, domain.ErrInternal)
	}

	// Only attempt to link an existing local account when the create failed
	// because of a duplicate login_id.
	if !errors.Is(createErr, domain.ErrDuplicateEntry) {
		return nil, fmt.Errorf("save new app user: %w: %w", createErr, domain.ErrInternal)
	}

	// Account linking: a user may already exist with the same login ID (e.g. password signup).
	existing, findErr := c.appUserByLoginIDFinder.FindByLoginID(ctx, orgID, loginID)
	if findErr != nil {
		if errors.Is(findErr, domain.ErrAppUserNotFound) {
			return nil, fmt.Errorf("save new app user: %w", createErr)
		}
		return nil, fmt.Errorf("find app user by login id for linking: %w", findErr)
	}

	// Refuse to auto-link an existing password-holding account.
	if existing.HashedPassword() != "" {
		return nil, fmt.Errorf("app user %s: %w", existing.ID().String(), domain.ErrAppUserAutoLinkRejected)
	}

	_, linkErr := domainuser.ProvisionProvider(ctx, c.providerSaver, existing.ID(), orgID, provider, providerID)
	if errors.Is(linkErr, domain.ErrDuplicateEntry) {
		return c.outputForLinkedProvider(ctx, orgID, provider, providerID, orgName)
	}
	if linkErr != nil {
		return nil, fmt.Errorf("link provider to app user %s: %w", existing.ID().String(), linkErr)
	}

	return c.buildOutput(existing.ID(), string(existing.LoginID()), orgName)
}

func (c *SupabaseExchangeCommand) linkNewUser(ctx context.Context, newUser *domainuser.AppUser, orgID domain.OrganizationID, provider string, providerID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	_, linkErr := domainuser.ProvisionProvider(ctx, c.providerSaver, newUser.ID(), orgID, provider, providerID)
	if linkErr != nil && !errors.Is(linkErr, domain.ErrDuplicateEntry) {
		return nil, fmt.Errorf("create provider link for new user: %w", linkErr)
	}

	c.publisher.Publish(domain.NewAppUserCreated(newUser.ID(), orgID, string(newUser.LoginID()), time.Now()))

	if linkErr != nil {
		return c.outputForLinkedProvider(ctx, orgID, provider, providerID, orgName)
	}
	return c.buildOutput(newUser.ID(), string(newUser.LoginID()), orgName)
}

// outputForLinkedProvider resolves the user of a provider link that a concurrent request created first.
func (c *SupabaseExchangeCommand) outputForLinkedProvider(ctx context.Context, orgID domain.OrganizationID, provider string, providerID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	providerLink, err := c.providerFinder.FindByProviderID(ctx, orgID, provider, providerID)
	if err != nil {
		return nil, fmt.Errorf("find provider link after duplicate: %w", err)
	}
	return c.outputForProviderLink(ctx, providerLink, orgName)
}

func (c *SupabaseExchangeCommand) outputForProviderLink(ctx context.Context, providerLink *domainuser.AppUserProvider, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	appUser, err := c.appUserByIDFinder.FindByID(ctx, providerLink.AppUserID())
	if err != nil {
		return nil, fmt.Errorf("find app user %s: %w", providerLink.AppUserID().String(), err)
	}
	return c.buildOutput(appUser.ID(), string(appUser.LoginID()), orgName)
}

func (c *SupabaseExchangeCommand) buildOutput(userID domain.AppUserID, loginID string, orgName string) (*authservice.SupabaseExchangeOutput, error) {
	output, err := authservice.NewSupabaseExchangeOutput(userID, loginID, orgName)
	if err != nil {
		return nil, fmt.Errorf("new supabase exchange output: %w", err)
	}

	return output, nil
}
