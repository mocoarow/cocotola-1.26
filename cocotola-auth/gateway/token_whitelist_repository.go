package gateway

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// --- access token whitelist ---

type accessTokenWhitelistRecord struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
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
func (r *AccessTokenWhitelistRepository) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	return findAndConvertWhitelist(ctx, r.db, userID, func(rec accessTokenWhitelistRecord) domain.WhitelistEntry {
		return domain.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}, "access token whitelist entries")
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *AccessTokenWhitelistRepository) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return saveWhitelist(ctx, r.db, whitelist, func(userID int, e domain.WhitelistEntry) accessTokenWhitelistRecord {
		return accessTokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}, "access token whitelist entries")
}

// --- refresh token whitelist ---

type refreshTokenWhitelistRecord struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
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
func (r *RefreshTokenWhitelistRepository) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	return findAndConvertWhitelist(ctx, r.db, userID, func(rec refreshTokenWhitelistRecord) domain.WhitelistEntry {
		return domain.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}, "refresh token whitelist entries")
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *RefreshTokenWhitelistRepository) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return saveWhitelist(ctx, r.db, whitelist, func(userID int, e domain.WhitelistEntry) refreshTokenWhitelistRecord {
		return refreshTokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}, "refresh token whitelist entries")
}

// --- session token whitelist ---

type sessionTokenWhitelistRecord struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
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
func (r *SessionTokenWhitelistRepository) FindByUserID(ctx context.Context, userID int) ([]domain.WhitelistEntry, error) {
	return findAndConvertWhitelist(ctx, r.db, userID, func(rec sessionTokenWhitelistRecord) domain.WhitelistEntry {
		return domain.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}, "session token whitelist entries")
}

// Save persists the whitelist aggregate by replacing all entries for the user.
func (r *SessionTokenWhitelistRepository) Save(ctx context.Context, whitelist *domain.TokenWhitelist) error {
	return saveWhitelist(ctx, r.db, whitelist, func(userID int, e domain.WhitelistEntry) sessionTokenWhitelistRecord {
		return sessionTokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}, "session token whitelist entries")
}
