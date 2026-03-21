package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type userNSpaceRecord struct {
	ID             int       `gorm:"column:id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	CreatedBy      int       `gorm:"column:created_by;<-:create"`
	OrganizationID int       `gorm:"column:organization_id"`
	UserID         int       `gorm:"column:user_id"`
	SpaceID        int       `gorm:"column:space_id"`
}

func (userNSpaceRecord) TableName() string {
	return "user_n_space"
}

// UserNSpaceRepository implements user-space association persistence using MySQL via GORM.
type UserNSpaceRepository struct {
	db *gorm.DB
}

// NewUserNSpaceRepository returns a new UserNSpaceRepository.
func NewUserNSpaceRepository(db *gorm.DB) *UserNSpaceRepository {
	return &UserNSpaceRepository{db: db}
}

// AddUserToSpace associates a user with a space.
func (r *UserNSpaceRepository) AddUserToSpace(ctx context.Context, organizationID int, userID int, spaceID int, createdBy int) error {
	record := userNSpaceRecord{
		ID:             0,
		CreatedAt:      time.Time{},
		CreatedBy:      createdBy,
		OrganizationID: organizationID,
		UserID:         userID,
		SpaceID:        spaceID,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("add user to space: %w", err)
	}
	return nil
}

// RemoveUserFromSpace removes the association between a user and a space.
func (r *UserNSpaceRepository) RemoveUserFromSpace(ctx context.Context, organizationID int, userID int, spaceID int) error {
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ? AND space_id = ?", organizationID, userID, spaceID).
		Delete(&userNSpaceRecord{
			ID:             0,
			CreatedAt:      time.Time{},
			CreatedBy:      0,
			OrganizationID: 0,
			UserID:         0,
			SpaceID:        0,
		}).Error; err != nil {
		return fmt.Errorf("remove user from space: %w", err)
	}
	return nil
}

// FindSpaceIDsByUserID returns the space IDs that the user is associated with.
func (r *UserNSpaceRepository) FindSpaceIDsByUserID(ctx context.Context, organizationID int, userID int) ([]int, error) {
	var records []userNSpaceRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find space ids by user id: %w", err)
	}
	ids := make([]int, len(records))
	for i, rec := range records {
		ids[i] = rec.SpaceID
	}
	return ids, nil
}

// IsUserInSpace checks whether a user is associated with a space.
func (r *UserNSpaceRepository) IsUserInSpace(ctx context.Context, organizationID int, userID int, spaceID int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&userNSpaceRecord{
		ID:             0,
		CreatedAt:      time.Time{},
		CreatedBy:      0,
		OrganizationID: 0,
		UserID:         0,
		SpaceID:        0,
	}).
		Where("organization_id = ? AND user_id = ? AND space_id = ?", organizationID, userID, spaceID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check user in space: %w", err)
	}
	return count > 0, nil
}
