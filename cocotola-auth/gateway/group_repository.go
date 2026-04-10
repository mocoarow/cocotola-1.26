package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

type groupRecord struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Version        int       `gorm:"column:version;->"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	UpdatedAt      time.Time `gorm:"column:updated_at;->"`
	CreatedBy      string    `gorm:"column:created_by;<-:create"`
	UpdatedBy      string    `gorm:"column:updated_by"`
	OrganizationID string    `gorm:"column:organization_id"`
	Name           string    `gorm:"column:name"`
	Enabled        bool      `gorm:"column:enabled"`
}

func (groupRecord) TableName() string {
	return "group"
}

func toGroupDomain(r *groupRecord) *domaingroup.Group {
	return domaingroup.ReconstructGroup(domain.MustParseGroupID(r.ID), domain.MustParseOrganizationID(r.OrganizationID), r.Name, r.Enabled)
}

// GroupRepository implements group persistence using GORM.
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository returns a new GroupRepository.
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Save persists a group record (upsert: insert or update).
func (r *GroupRepository) Save(ctx context.Context, group *domaingroup.Group) error {
	systemUserID := domain.SystemAppUserID().String()
	record := groupRecord{
		ID:             group.ID().String(),
		Version:        0,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      systemUserID,
		UpdatedBy:      systemUserID,
		OrganizationID: group.OrganizationID().String(),
		Name:           group.Name(),
		Enabled:        group.Enabled(),
	}
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save group: %w", err)
	}
	return nil
}

// Create inserts a new group record and returns the generated ID.
func (r *GroupRepository) Create(ctx context.Context, organizationID domain.OrganizationID, name string) (domain.GroupID, error) {
	groupID, err := domain.NewGroupIDV7()
	if err != nil {
		return domain.GroupID{}, fmt.Errorf("generate group id: %w", err)
	}
	systemUserID := domain.SystemAppUserID().String()
	record := groupRecord{
		ID:             groupID.String(),
		Version:        0,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      systemUserID,
		UpdatedBy:      systemUserID,
		OrganizationID: organizationID.String(),
		Name:           name,
		Enabled:        true,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return domain.GroupID{}, fmt.Errorf("create group: %w", err)
	}
	return groupID, nil
}

// FindByID looks up a group by its ID.
func (r *GroupRepository) FindByID(ctx context.Context, id domain.GroupID) (*domaingroup.Group, error) {
	var record groupRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("find group by id: %w", err)
	}
	return toGroupDomain(&record), nil
}

// FindByName looks up a group by organization ID and name.
func (r *GroupRepository) FindByName(ctx context.Context, organizationID domain.OrganizationID, name string) (*domaingroup.Group, error) {
	var record groupRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND name = ?", organizationID.String(), name).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("find group by name: %w", err)
	}
	return toGroupDomain(&record), nil
}
