package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/gormsave"
)

type organizationRecord struct {
	ID              string    `gorm:"column:id;primaryKey"`
	Version         int       `gorm:"column:version"`
	CreatedAt       time.Time `gorm:"column:created_at;->"`
	UpdatedAt       time.Time `gorm:"column:updated_at;->"`
	CreatedBy       string    `gorm:"column:created_by;<-:create"`
	UpdatedBy       string    `gorm:"column:updated_by"`
	Name            string    `gorm:"column:name"`
	MaxActiveUsers  int       `gorm:"column:max_active_users"`
	MaxActiveGroups int       `gorm:"column:max_active_groups"`
}

func (organizationRecord) TableName() string {
	return "organization"
}

func (r *organizationRecord) GetVersion() int {
	return r.Version
}

func toOrganizationDomain(r *organizationRecord) *domain.Organization {
	o := domain.ReconstructOrganization(domain.MustParseOrganizationID(r.ID), r.Name, r.MaxActiveUsers, r.MaxActiveGroups)
	o.SetVersion(r.Version)
	return o
}

// OrganizationRepository implements organization persistence using GORM.
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository returns a new OrganizationRepository.
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Save persists an organization aggregate. New aggregates (version 0) are
// inserted; loaded aggregates (version > 0) are updated via CAS on the version
// column. The repository updates the aggregate's version after a successful
// persist so the caller does not need to manage versioning.
func (r *OrganizationRepository) Save(ctx context.Context, org *domain.Organization) error {
	systemUserID := domain.SystemAppUserID().String()
	record := organizationRecord{
		ID:              org.ID().String(),
		Version:         org.Version() + 1,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
		CreatedBy:       systemUserID,
		UpdatedBy:       systemUserID,
		Name:            org.Name(),
		MaxActiveUsers:  org.MaxActiveUsers(),
		MaxActiveGroups: org.MaxActiveGroups(),
	}
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*organizationRecord]{
		DB:     r.db,
		Entity: org,
		Record: &record,
		PK:     map[string]any{"id": record.ID},
		Updates: map[string]any{
			colName:             record.Name,
			"max_active_users":  record.MaxActiveUsers,
			"max_active_groups": record.MaxActiveGroups,
		},
		EntityName:   "organization",
		OmitOnInsert: nil,
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return domain.ErrOrganizationNotFound
	}
	if err != nil {
		return fmt.Errorf("save organization: %w", err)
	}
	return nil
}

// FindByID looks up an organization by its ID.
func (r *OrganizationRepository) FindByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	var record organizationRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("find organization by id: %w", err)
	}
	return toOrganizationDomain(&record), nil
}

// FindByName looks up an organization by its name.
func (r *OrganizationRepository) FindByName(ctx context.Context, name string) (*domain.Organization, error) {
	var record organizationRecord
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("find organization by name: %w", err)
	}
	return toOrganizationDomain(&record), nil
}
