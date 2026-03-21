package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type groupRecord struct {
	ID             int       `gorm:"column:id;primaryKey"`
	Version        int       `gorm:"column:version;->"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	UpdatedAt      time.Time `gorm:"column:updated_at;->"`
	CreatedBy      int       `gorm:"column:created_by;<-:create"`
	UpdatedBy      int       `gorm:"column:updated_by"`
	OrganizationID int       `gorm:"column:organization_id"`
	Name           string    `gorm:"column:name"`
	Enabled        bool      `gorm:"column:enabled"`
}

func (groupRecord) TableName() string {
	return "group"
}

func toGroupDomain(r *groupRecord) *domain.Group {
	return domain.ReconstructGroup(r.ID, r.OrganizationID, r.Name, r.Enabled)
}

// GroupRepository implements group persistence using MySQL via GORM.
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository returns a new GroupRepository.
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Save persists a group record (upsert: insert or update).
func (r *GroupRepository) Save(ctx context.Context, group *domain.Group) error {
	record := groupRecord{
		ID:             group.ID(),
		Version:        0,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      0,
		UpdatedBy:      0,
		OrganizationID: group.OrganizationID(),
		Name:           group.Name(),
		Enabled:        group.Enabled(),
	}
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save group: %w", err)
	}
	return nil
}

// Create inserts a new group record and returns the auto-generated ID.
func (r *GroupRepository) Create(ctx context.Context, organizationID int, name string) (int, error) {
	record := groupRecord{
		ID:             0,
		Version:        0,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      0,
		UpdatedBy:      0,
		OrganizationID: organizationID,
		Name:           name,
		Enabled:        true,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return 0, fmt.Errorf("create group: %w", err)
	}
	return record.ID, nil
}

// FindByID looks up a group by its ID.
func (r *GroupRepository) FindByID(ctx context.Context, id int) (*domain.Group, error) {
	var record groupRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("find group by id: %w", err)
	}
	return toGroupDomain(&record), nil
}

// FindByName looks up a group by organization ID and name.
func (r *GroupRepository) FindByName(ctx context.Context, organizationID int, name string) (*domain.Group, error) {
	var record groupRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND name = ?", organizationID, name).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("find group by name: %w", err)
	}
	return toGroupDomain(&record), nil
}
