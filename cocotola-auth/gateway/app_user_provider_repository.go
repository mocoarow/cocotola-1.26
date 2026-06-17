package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/gormsave"
)

type appUserProviderRecord struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Version        int       `gorm:"column:version"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	UpdatedAt      time.Time `gorm:"column:updated_at;->"`
	CreatedBy      string    `gorm:"column:created_by;<-:create"`
	UpdatedBy      string    `gorm:"column:updated_by"`
	AppUserID      string    `gorm:"column:app_user_id"`
	OrganizationID string    `gorm:"column:organization_id"`
	Provider       string    `gorm:"column:provider"`
	ProviderID     string    `gorm:"column:provider_id"`
}

func (appUserProviderRecord) TableName() string {
	return "app_user_provider"
}

func (r *appUserProviderRecord) GetVersion() int {
	return r.Version
}

func toAppUserProviderDomain(r *appUserProviderRecord) (*domainuser.AppUserProvider, error) {
	id, err := domain.ParseAppUserProviderID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid app user provider id %q in db: %w", r.ID, err)
	}
	appUserID, err := domain.ParseAppUserID(r.AppUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid app user id %q in db: %w", r.AppUserID, err)
	}
	orgID, err := domain.ParseOrganizationID(r.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id %q in db: %w", r.OrganizationID, err)
	}
	p := domainuser.ReconstructAppUserProvider(id, appUserID, orgID, r.Provider, r.ProviderID)
	p.SetVersion(r.Version)
	return p, nil
}

// AppUserProviderRepository implements app user provider persistence using GORM.
type AppUserProviderRepository struct {
	db *gorm.DB
}

// NewAppUserProviderRepository returns a new AppUserProviderRepository.
func NewAppUserProviderRepository(db *gorm.DB) *AppUserProviderRepository {
	return &AppUserProviderRepository{db: db}
}

// Save persists an app user provider entity. New entities (version 0) are inserted;
// loaded entities (version > 0) are updated via CAS on the version column.
func (r *AppUserProviderRepository) Save(ctx context.Context, p *domainuser.AppUserProvider) error {
	systemUserID := domain.SystemAppUserID().String()
	record := appUserProviderRecord{
		ID:             p.ID().String(),
		Version:        p.Version() + 1,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      systemUserID,
		UpdatedBy:      systemUserID,
		AppUserID:      p.AppUserID().String(),
		OrganizationID: p.OrganizationID().String(),
		Provider:       p.Provider(),
		ProviderID:     p.ProviderID(),
	}
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*appUserProviderRecord]{
		DB:     r.db,
		Entity: p,
		Record: &record,
		PK:     map[string]any{"id": record.ID},
		Updates: map[string]any{
			"app_user_id":     record.AppUserID,
			colOrganizationID: record.OrganizationID,
			"provider":        record.Provider,
			"provider_id":     record.ProviderID,
		},
		EntityName:   "app user provider",
		OmitOnInsert: nil,
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return domain.ErrAppUserProviderNotFound
	}
	if err != nil {
		return fmt.Errorf("save app user provider: %w", err)
	}
	return nil
}

// FindByProviderID looks up an app user provider link by organization, provider, and provider ID.
func (r *AppUserProviderRepository) FindByProviderID(ctx context.Context, organizationID domain.OrganizationID, provider string, providerID string) (*domainuser.AppUserProvider, error) {
	var record appUserProviderRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND provider = ? AND provider_id = ?", organizationID.String(), provider, providerID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserProviderNotFound
		}
		return nil, fmt.Errorf("find app user provider by provider id: %w", err)
	}
	p, err := toAppUserProviderDomain(&record)
	if err != nil {
		return nil, fmt.Errorf("convert app user provider domain: %w", err)
	}
	return p, nil
}

// FindByAppUserID looks up all provider links for a given app user.
func (r *AppUserProviderRepository) FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) ([]domainuser.AppUserProvider, error) {
	var records []appUserProviderRecord
	if err := r.db.WithContext(ctx).
		Where("app_user_id = ?", appUserID.String()).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find app user providers by app user id: %w", err)
	}
	result := make([]domainuser.AppUserProvider, len(records))
	for i := range records {
		p, err := toAppUserProviderDomain(&records[i])
		if err != nil {
			return nil, fmt.Errorf("convert app user provider domain: %w", err)
		}
		result[i] = *p
	}
	return result, nil
}
