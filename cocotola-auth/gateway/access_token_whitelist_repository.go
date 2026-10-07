package gateway

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

type accessTokenWhitelistRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (accessTokenWhitelistRecord) TableName() string { return "access_token_whitelist" }

// AccessTokenWhitelistRepository implements whitelist persistence for access tokens.
type AccessTokenWhitelistRepository struct{ db *gorm.DB }

// NewAccessTokenWhitelistRepository returns a new AccessTokenWhitelistRepository.
func NewAccessTokenWhitelistRepository(db *gorm.DB) *AccessTokenWhitelistRepository {
	return &AccessTokenWhitelistRepository{db: db}
}

// FindByUserID returns all whitelist entries for the given user.
func (r *AccessTokenWhitelistRepository) FindByUserID(ctx context.Context, userID domain.AppUserID) ([]domaintoken.WhitelistEntry, error) {
	return findAndConvertWhitelist(ctx, r.db, userID, func(rec accessTokenWhitelistRecord) domaintoken.WhitelistEntry {
		return domaintoken.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}, "access token whitelist entries")
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *AccessTokenWhitelistRepository) Save(ctx context.Context, whitelist *domaintoken.Whitelist) error {
	return saveWhitelist(ctx, r.db, whitelist, func(userID string, e domaintoken.WhitelistEntry) accessTokenWhitelistRecord {
		return accessTokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}, "access token whitelist entries")
}
