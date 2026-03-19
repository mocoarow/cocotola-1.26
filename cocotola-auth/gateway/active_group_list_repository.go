package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type activeGroupRecord struct {
	OrganizationID int       `gorm:"column:organization_id;primaryKey"`
	GroupID        int       `gorm:"column:group_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (activeGroupRecord) TableName() string {
	return "active_group"
}

// ActiveGroupListRepository implements active group list persistence using MySQL via GORM.
type ActiveGroupListRepository struct {
	db *gorm.DB
}

// NewActiveGroupListRepository returns a new ActiveGroupListRepository.
func NewActiveGroupListRepository(db *gorm.DB) *ActiveGroupListRepository {
	return &ActiveGroupListRepository{db: db}
}

// FindByOrganizationID returns the active group list for the given organization.
func (r *ActiveGroupListRepository) FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveGroupList, error) {
	var records []activeGroupRecord
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find active groups by organization id: %w", err)
	}

	groupIDs := make([]int, len(records))
	for i := range records {
		groupIDs[i] = records[i].GroupID
	}

	list, err := domain.NewActiveGroupList(organizationID, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active group list: %w", err)
	}
	return list, nil
}

// Save persists the active group list by replacing all entries for the organization.
func (r *ActiveGroupListRepository) Save(ctx context.Context, list *domain.ActiveGroupList) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ?", list.OrganizationID()).
			Delete(&activeGroupRecord{}).Error; err != nil {
			return fmt.Errorf("delete active group entries: %w", err)
		}

		entries := list.Entries()
		if len(entries) == 0 {
			return nil
		}

		records := make([]activeGroupRecord, len(entries))
		for i, groupID := range entries {
			records[i] = activeGroupRecord{
				OrganizationID: list.OrganizationID(),
				GroupID:        groupID,
				CreatedAt:      time.Now(),
			}
		}

		if err := tx.Create(&records).Error; err != nil {
			return fmt.Errorf("insert active group entries: %w", err)
		}
		return nil
	})
}
