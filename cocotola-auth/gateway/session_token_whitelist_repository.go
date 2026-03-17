package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type sessionTokenWhitelistRecord struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (sessionTokenWhitelistRecord) TableName() string {
	return "session_token_whitelist"
}

// SessionTokenWhitelistRepository implements whitelist persistence for session tokens.
type SessionTokenWhitelistRepository struct {
	db *gorm.DB
}

// NewSessionTokenWhitelistRepository returns a new SessionTokenWhitelistRepository.
func NewSessionTokenWhitelistRepository(db *gorm.DB) *SessionTokenWhitelistRepository {
	return &SessionTokenWhitelistRepository{db: db}
}

// FindByUserID returns all whitelist entries for the given user.
func (r *SessionTokenWhitelistRepository) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	var records []sessionTokenWhitelistRecord
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find session token whitelist entries: %w", err)
	}

	entries := make([]domain.WhitelistEntry, len(records))
	for i := range records {
		entries[i] = domain.WhitelistEntry{ID: records[i].TokenID, CreatedAt: records[i].CreatedAt}
	}
	return entries, nil
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *SessionTokenWhitelistRepository) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", whitelist.UserID()).
			Delete(&sessionTokenWhitelistRecord{}).Error; err != nil {
			return fmt.Errorf("delete session token whitelist entries: %w", err)
		}

		entries := whitelist.Entries()
		if len(entries) == 0 {
			return nil
		}

		records := make([]sessionTokenWhitelistRecord, len(entries))
		for i, e := range entries {
			records[i] = sessionTokenWhitelistRecord{
				UserID:    whitelist.UserID(),
				TokenID:   e.ID,
				CreatedAt: e.CreatedAt,
			}
		}

		if err := tx.Create(&records).Error; err != nil {
			return fmt.Errorf("insert session token whitelist entries: %w", err)
		}
		return nil
	})
}
