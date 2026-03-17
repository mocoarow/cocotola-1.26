package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type refreshTokenWhitelistRecord struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (refreshTokenWhitelistRecord) TableName() string {
	return "refresh_token_whitelist"
}

// RefreshTokenWhitelistRepository implements whitelist persistence for refresh tokens.
type RefreshTokenWhitelistRepository struct {
	db *gorm.DB
}

// NewRefreshTokenWhitelistRepository returns a new RefreshTokenWhitelistRepository.
func NewRefreshTokenWhitelistRepository(db *gorm.DB) *RefreshTokenWhitelistRepository {
	return &RefreshTokenWhitelistRepository{db: db}
}

// FindByUserID returns all whitelist entries for the given user.
func (r *RefreshTokenWhitelistRepository) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	var records []refreshTokenWhitelistRecord
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find refresh token whitelist entries: %w", err)
	}

	entries := make([]domain.WhitelistEntry, len(records))
	for i := range records {
		entries[i] = domain.WhitelistEntry{ID: records[i].TokenID, CreatedAt: records[i].CreatedAt}
	}
	return entries, nil
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *RefreshTokenWhitelistRepository) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", whitelist.UserID()).
			Delete(&refreshTokenWhitelistRecord{}).Error; err != nil {
			return fmt.Errorf("delete refresh token whitelist entries: %w", err)
		}

		entries := whitelist.Entries()
		if len(entries) == 0 {
			return nil
		}

		records := make([]refreshTokenWhitelistRecord, len(entries))
		for i, e := range entries {
			records[i] = refreshTokenWhitelistRecord{
				UserID:    whitelist.UserID(),
				TokenID:   e.ID,
				CreatedAt: e.CreatedAt,
			}
		}

		if err := tx.Create(&records).Error; err != nil {
			return fmt.Errorf("insert refresh token whitelist entries: %w", err)
		}
		return nil
	})
}
