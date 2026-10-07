package gateway

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

type sessionTokenWhitelistRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (sessionTokenWhitelistRecord) TableName() string { return "session_token_whitelist" }

// SessionTokenWhitelistRepository implements whitelist persistence for session tokens.
type SessionTokenWhitelistRepository struct{ db *gorm.DB }

// NewSessionTokenWhitelistRepository returns a new SessionTokenWhitelistRepository.
func NewSessionTokenWhitelistRepository(db *gorm.DB) *SessionTokenWhitelistRepository {
	return &SessionTokenWhitelistRepository{db: db}
}

// FindByUserID returns all whitelist entries for the given user.
func (r *SessionTokenWhitelistRepository) FindByUserID(ctx context.Context, userID domain.AppUserID) ([]domaintoken.WhitelistEntry, error) {
	return findAndConvertWhitelist(ctx, r.db, userID, func(rec sessionTokenWhitelistRecord) domaintoken.WhitelistEntry {
		return domaintoken.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}, "session token whitelist entries")
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *SessionTokenWhitelistRepository) Save(ctx context.Context, whitelist *domaintoken.Whitelist) error {
	return saveWhitelist(ctx, r.db, whitelist, func(userID string, e domaintoken.WhitelistEntry) sessionTokenWhitelistRecord {
		return sessionTokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}, "session token whitelist entries")
}
