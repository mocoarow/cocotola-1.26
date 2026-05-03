package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type userSettingRecord struct {
	AppUserID    string    `gorm:"column:app_user_id;primaryKey"`
	Version      int       `gorm:"column:version"`
	CreatedAt    time.Time `gorm:"column:created_at;->"`
	UpdatedAt    time.Time `gorm:"column:updated_at;->"`
	CreatedBy    string    `gorm:"column:created_by;<-:create"`
	UpdatedBy    string    `gorm:"column:updated_by"`
	MaxWorkbooks int       `gorm:"column:max_workbooks"`
	Language     string    `gorm:"column:language"`
}

func (userSettingRecord) TableName() string {
	return "user_setting"
}

func toUserSettingDomain(r *userSettingRecord) (*domain.UserSetting, error) {
	appUserID, err := domain.ParseAppUserID(r.AppUserID)
	if err != nil {
		return nil, fmt.Errorf("parse app user id %s: %w", r.AppUserID, err)
	}
	setting, err := domain.ReconstructUserSetting(appUserID, r.MaxWorkbooks, r.Language)
	if err != nil {
		return nil, fmt.Errorf("reconstruct user setting: %w", err)
	}
	setting.SetVersion(r.Version)
	return setting, nil
}

// UserSettingRepository implements user setting persistence using GORM.
type UserSettingRepository struct {
	db *gorm.DB
}

// NewUserSettingRepository returns a new UserSettingRepository.
func NewUserSettingRepository(db *gorm.DB) *UserSettingRepository {
	return &UserSettingRepository{db: db}
}

// Save persists a user setting. New settings (version 0) are inserted;
// loaded settings (version > 0) are updated via CAS on the version column.
func (r *UserSettingRepository) Save(ctx context.Context, setting *domain.UserSetting) error {
	nextVersion := setting.Version() + 1
	operatorID := setting.AppUserID().String()
	record := userSettingRecord{
		AppUserID:    setting.AppUserID().String(),
		Version:      nextVersion,
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
		CreatedBy:    operatorID,
		UpdatedBy:    operatorID,
		MaxWorkbooks: setting.MaxWorkbooks(),
		Language:     setting.Language(),
	}
	if setting.Version() == 0 {
		if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
			return fmt.Errorf("insert user setting: %w", err)
		}
		setting.SetVersion(nextVersion)
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&record).
		Where("app_user_id = ? AND version = ?", record.AppUserID, setting.Version()).
		Updates(map[string]any{
			"max_workbooks": record.MaxWorkbooks,
			"language":      record.Language,
			colVersion:      nextVersion,
			"updated_by":    operatorID,
		})
	if result.Error != nil {
		return fmt.Errorf("update user setting: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserSettingConcurrentModification
	}
	setting.SetVersion(nextVersion)
	return nil
}

// FindByAppUserID looks up a user setting by the app user ID.
func (r *UserSettingRepository) FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error) {
	var record userSettingRecord
	if err := r.db.WithContext(ctx).Where("app_user_id = ?", appUserID.String()).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserSettingNotFound
		}
		return nil, fmt.Errorf("find user setting by app user id: %w", err)
	}
	setting, err := toUserSettingDomain(&record)
	if err != nil {
		return nil, fmt.Errorf("find user setting by app user id: %w", err)
	}
	return setting, nil
}
