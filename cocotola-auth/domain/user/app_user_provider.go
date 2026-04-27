package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// AppUserProvider represents a link between an AppUser and an external identity provider.
// One AppUser can have multiple provider links (e.g. Google, Facebook, LINE).
type AppUserProvider struct {
	id             domain.AppUserProviderID
	version        int
	appUserID      domain.AppUserID
	organizationID domain.OrganizationID
	provider       string
	providerID     string
}

// NewAppUserProvider creates a validated AppUserProvider for a brand-new aggregate (version 0).
func NewAppUserProvider(id domain.AppUserProviderID, appUserID domain.AppUserID, organizationID domain.OrganizationID, provider string, providerID string) (*AppUserProvider, error) {
	m := &AppUserProvider{
		id:             id,
		version:        0,
		appUserID:      appUserID,
		organizationID: organizationID,
		provider:       provider,
		providerID:     providerID,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructAppUserProvider reconstitutes an AppUserProvider from persistence.
func ReconstructAppUserProvider(id domain.AppUserProviderID, appUserID domain.AppUserID, organizationID domain.OrganizationID, provider string, providerID string) *AppUserProvider {
	return &AppUserProvider{
		id:             id,
		version:        0,
		appUserID:      appUserID,
		organizationID: organizationID,
		provider:       provider,
		providerID:     providerID,
	}
}

// SetVersion sets the persisted row version.
// Only the gateway/repository layer should call this, after a successful load or save.
func (p *AppUserProvider) SetVersion(version int) {
	p.version = version
}

// Version returns the aggregate version (0 = new, not yet saved).
func (p *AppUserProvider) Version() int { return p.version }

// ID returns the provider link ID.
func (p *AppUserProvider) ID() domain.AppUserProviderID { return p.id }

// AppUserID returns the linked app user ID.
func (p *AppUserProvider) AppUserID() domain.AppUserID { return p.appUserID }

// OrganizationID returns the organization ID.
func (p *AppUserProvider) OrganizationID() domain.OrganizationID { return p.organizationID }

// Provider returns the external identity provider name.
func (p *AppUserProvider) Provider() string { return p.provider }

// ProviderID returns the external identity provider user ID.
func (p *AppUserProvider) ProviderID() string { return p.providerID }

func (p *AppUserProvider) validate() error {
	if p.id.IsZero() {
		return errors.New("app user provider id must not be zero")
	}
	if p.appUserID.IsZero() {
		return errors.New("app user provider app user id must not be zero")
	}
	if p.organizationID.IsZero() {
		return errors.New("app user provider organization id must not be zero")
	}
	if p.provider == "" {
		return errors.New("app user provider: name is required")
	}
	if p.providerID == "" {
		return errors.New("app user provider: external id is required")
	}
	return nil
}

// AppUserProviderSaver persists an AppUserProvider entity.
type AppUserProviderSaver interface {
	Save(ctx context.Context, provider *AppUserProvider) error
}

// ProvisionProvider generates a fresh UUIDv7 ID, constructs an AppUserProvider
// via the domain factory, and persists it.
func ProvisionProvider(
	ctx context.Context,
	saver AppUserProviderSaver,
	appUserID domain.AppUserID,
	organizationID domain.OrganizationID,
	provider string,
	providerID string,
) (*AppUserProvider, error) {
	id, err := domain.NewAppUserProviderIDV7()
	if err != nil {
		return nil, fmt.Errorf("generate app user provider id: %w", err)
	}
	p, err := NewAppUserProvider(id, appUserID, organizationID, provider, providerID)
	if err != nil {
		return nil, fmt.Errorf("new app user provider: %w", err)
	}
	if err := saver.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("save app user provider: %w", err)
	}
	return p, nil
}
