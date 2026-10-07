package gateway

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

type refreshTokenWhitelistRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (refreshTokenWhitelistRecord) TableName() string { return "refresh_token_whitelist" }

// RefreshTokenWhitelistRepository implements whitelist persistence for refresh tokens.
type RefreshTokenWhitelistRepository struct{ db *gorm.DB }

// NewRefreshTokenWhitelistRepository returns a new RefreshTokenWhitelistRepository.
func NewRefreshTokenWhitelistRepository(db *gorm.DB) *RefreshTokenWhitelistRepository {
	return &RefreshTokenWhitelistRepository{db: db}
}

// FindByUserID returns all whitelist entries for the given user.
func (r *RefreshTokenWhitelistRepository) FindByUserID(ctx context.Context, userID domain.AppUserID) ([]domaintoken.WhitelistEntry, error) {
	return findAndConvertWhitelist(ctx, r.db, userID, func(rec refreshTokenWhitelistRecord) domaintoken.WhitelistEntry {
		return domaintoken.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}, "refresh token whitelist entries")
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *RefreshTokenWhitelistRepository) Save(ctx context.Context, whitelist *domaintoken.Whitelist) error {
	return saveWhitelist(ctx, r.db, whitelist, func(userID string, e domaintoken.WhitelistEntry) refreshTokenWhitelistRecord {
		return refreshTokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}, "refresh token whitelist entries")
}
