package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
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

func toAppUserProviderDomain(r *appUserProviderRecord) *domainuser.AppUserProvider {
	p := domainuser.ReconstructAppUserProvider(
		domain.MustParseAppUserProviderID(r.ID),
		domain.MustParseAppUserID(r.AppUserID),
		domain.MustParseOrganizationID(r.OrganizationID),
		r.Provider,
		r.ProviderID,
	)
	p.SetVersion(r.Version)
	return p
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
	nextVersion := p.Version() + 1
	systemUserID := domain.SystemAppUserID().String()
	record := appUserProviderRecord{
		ID:             p.ID().String(),
		Version:        nextVersion,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      systemUserID,
		UpdatedBy:      systemUserID,
		AppUserID:      p.AppUserID().String(),
		OrganizationID: p.OrganizationID().String(),
		Provider:       p.Provider(),
		ProviderID:     p.ProviderID(),
	}
	if p.Version() == 0 {
		if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
			return fmt.Errorf("insert app user provider: %w", err)
		}
		p.SetVersion(nextVersion)
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&record).
		Where("id = ? AND version = ?", record.ID, p.Version()).
		Updates(map[string]any{
			"app_user_id":     record.AppUserID,
			"organization_id": record.OrganizationID,
			"provider":        record.Provider,
			"provider_id":     record.ProviderID,
			"version":         nextVersion,
		})
	if result.Error != nil {
		return fmt.Errorf("update app user provider: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrAppUserProviderConcurrentModification
	}
	p.SetVersion(nextVersion)
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
	return toAppUserProviderDomain(&record), nil
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
		result[i] = *toAppUserProviderDomain(&records[i])
	}
	return result, nil
}
