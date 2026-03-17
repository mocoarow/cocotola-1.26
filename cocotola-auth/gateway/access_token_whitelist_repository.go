package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type accessTokenWhitelistRecord struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (accessTokenWhitelistRecord) TableName() string {
	return "access_token_whitelist"
}

// AccessTokenWhitelistRepository implements whitelist persistence for access tokens.
type AccessTokenWhitelistRepository struct {
	db *gorm.DB
}

// NewAccessTokenWhitelistRepository returns a new AccessTokenWhitelistRepository.
func NewAccessTokenWhitelistRepository(db *gorm.DB) *AccessTokenWhitelistRepository {
	return &AccessTokenWhitelistRepository{db: db}
}

// FindByUserID returns all whitelist entries for the given user.
func (r *AccessTokenWhitelistRepository) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	var records []accessTokenWhitelistRecord
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find access token whitelist entries: %w", err)
	}

	entries := make([]domain.WhitelistEntry, len(records))
	for i := range records {
		entries[i] = domain.WhitelistEntry{ID: records[i].TokenID, CreatedAt: records[i].CreatedAt}
	}
	return entries, nil
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *AccessTokenWhitelistRepository) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", whitelist.UserID()).
			Delete(&accessTokenWhitelistRecord{}).Error; err != nil {
			return fmt.Errorf("delete access token whitelist entries: %w", err)
		}

		entries := whitelist.Entries()
		if len(entries) == 0 {
			return nil
		}

		records := make([]accessTokenWhitelistRecord, len(entries))
		for i, e := range entries {
			records[i] = accessTokenWhitelistRecord{
				UserID:    whitelist.UserID(),
				TokenID:   e.ID,
				CreatedAt: e.CreatedAt,
			}
		}

		if err := tx.Create(&records).Error; err != nil {
			return fmt.Errorf("insert access token whitelist entries: %w", err)
		}
		return nil
	})
}
