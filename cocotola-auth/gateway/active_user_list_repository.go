package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type activeUserRecord struct {
	OrganizationID int       `gorm:"column:organization_id;primaryKey"`
	UserID         int       `gorm:"column:user_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (activeUserRecord) TableName() string {
	return "active_user"
}

// ActiveUserListRepository implements active user list persistence using MySQL via GORM.
type ActiveUserListRepository struct {
	db *gorm.DB
}

// NewActiveUserListRepository returns a new ActiveUserListRepository.
func NewActiveUserListRepository(db *gorm.DB) *ActiveUserListRepository {
	return &ActiveUserListRepository{db: db}
}

// FindByOrganizationID returns the active user list for the given organization.
func (r *ActiveUserListRepository) FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveUserList, error) {
	var records []activeUserRecord
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find active users by organization id: %w", err)
	}

	userIDs := make([]int, len(records))
	for i := range records {
		userIDs[i] = records[i].UserID
	}

	list, err := domain.NewActiveUserList(organizationID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active user list: %w", err)
	}
	return list, nil
}

// Save persists the active user list by replacing all entries for the organization.
func (r *ActiveUserListRepository) Save(ctx context.Context, list *domain.ActiveUserList) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ?", list.OrganizationID()).
			Delete(&activeUserRecord{}).Error; err != nil {
			return fmt.Errorf("delete active user entries: %w", err)
		}

		entries := list.Entries()
		if len(entries) == 0 {
			return nil
		}

		records := make([]activeUserRecord, len(entries))
		for i, userID := range entries {
			records[i] = activeUserRecord{
				OrganizationID: list.OrganizationID(),
				UserID:         userID,
				CreatedAt:      time.Now(),
			}
		}

		if err := tx.Create(&records).Error; err != nil {
			return fmt.Errorf("insert active user entries: %w", err)
		}
		return nil
	})
}
