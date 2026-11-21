package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/gormsave"
)

type tokenWhitelistRootRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	Version   int       `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at;->"`
	UpdatedAt time.Time `gorm:"column:updated_at;->"`
}

func (r *tokenWhitelistRootRecord) GetVersion() int { return r.Version }

type tokenWhitelistRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	TokenID   string    `gorm:"column:token_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type tokenWhitelistRepository struct {
	db        *gorm.DB
	rootTable string
	table     string
}

// AccessTokenWhitelistRepository implements whitelist persistence for access tokens.
type AccessTokenWhitelistRepository struct{ tokenWhitelistRepository }

// NewAccessTokenWhitelistRepository returns a new AccessTokenWhitelistRepository.
func NewAccessTokenWhitelistRepository(db *gorm.DB) *AccessTokenWhitelistRepository {
	return &AccessTokenWhitelistRepository{tokenWhitelistRepository{db: db, rootTable: "access_token_whitelist_root", table: "access_token_whitelist"}}
}

// RefreshTokenWhitelistRepository implements whitelist persistence for refresh tokens.
type RefreshTokenWhitelistRepository struct{ tokenWhitelistRepository }

// NewRefreshTokenWhitelistRepository returns a new RefreshTokenWhitelistRepository.
func NewRefreshTokenWhitelistRepository(db *gorm.DB) *RefreshTokenWhitelistRepository {
	return &RefreshTokenWhitelistRepository{tokenWhitelistRepository{db: db, rootTable: "refresh_token_whitelist_root", table: "refresh_token_whitelist"}}
}

// SessionTokenWhitelistRepository implements whitelist persistence for session tokens.
type SessionTokenWhitelistRepository struct{ tokenWhitelistRepository }

// NewSessionTokenWhitelistRepository returns a new SessionTokenWhitelistRepository.
func NewSessionTokenWhitelistRepository(db *gorm.DB) *SessionTokenWhitelistRepository {
	return &SessionTokenWhitelistRepository{tokenWhitelistRepository{db: db, rootTable: "session_token_whitelist_root", table: "session_token_whitelist"}}
}

// FindByUserID returns the whitelist of the given user. A user without a saved whitelist gets an empty one at version 0.
func (r *tokenWhitelistRepository) FindByUserID(ctx context.Context, userID domain.AppUserID) (*domaintoken.Whitelist, error) {
	db := r.db.WithContext(ctx)
	var roots []tokenWhitelistRootRecord
	if err := db.Table(r.rootTable).Where("user_id = ?", userID.String()).Find(&roots).Error; err != nil {
		return nil, fmt.Errorf("find %s of user %s: %w", r.rootTable, userID, err)
	}
	var records []tokenWhitelistRecord
	if err := db.Table(r.table).Where("user_id = ?", userID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s of user %s: %w", r.table, userID, err)
	}

	entries := make([]domaintoken.WhitelistEntry, len(records))
	for i, rec := range records {
		entries[i] = domaintoken.WhitelistEntry{ID: rec.TokenID, CreatedAt: rec.CreatedAt}
	}
	whitelist, err := domaintoken.NewWhitelist(userID, entries)
	if err != nil {
		return nil, fmt.Errorf("reconstruct whitelist of user %s: %w", userID, err)
	}
	if len(roots) > 0 {
		whitelist.SetVersion(roots[0].Version)
	}
	return whitelist, nil
}

// Save persists the whitelist with an optimistic lock on its root row.
func (r *tokenWhitelistRepository) Save(ctx context.Context, whitelist *domaintoken.Whitelist) error {
	userID := whitelist.UserID().String()
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		root := tokenWhitelistRootRecord{UserID: userID, Version: whitelist.Version() + 1, CreatedAt: time.Time{}, UpdatedAt: time.Time{}}
		err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*tokenWhitelistRootRecord]{
			DB:           tx.Table(r.rootTable),
			Entity:       whitelist,
			Record:       &root,
			PK:           map[string]any{"user_id": userID},
			Updates:      map[string]any{"updated_at": gorm.Expr("CURRENT_TIMESTAMP")},
			EntityName:   r.rootTable,
			OmitOnInsert: nil,
		})
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("insert %s: %w", r.rootTable, libversioned.ErrConcurrentModification)
		}
		if err != nil {
			return fmt.Errorf("save %s: %w", r.rootTable, err)
		}
		return r.replaceEntries(tx, userID, whitelist.Entries())
	}); err != nil {
		return fmt.Errorf("save whitelist: %w", err)
	}
	return nil
}

func (r *tokenWhitelistRepository) replaceEntries(tx *gorm.DB, userID string, entries []domaintoken.WhitelistEntry) error {
	if err := tx.Table(r.table).Where("user_id = ?", userID).Delete(new(tokenWhitelistRecord)).Error; err != nil {
		return fmt.Errorf("delete %s: %w", r.table, err)
	}
	if len(entries) == 0 {
		return nil
	}
	records := make([]tokenWhitelistRecord, len(entries))
	for i, e := range entries {
		records[i] = tokenWhitelistRecord{UserID: userID, TokenID: e.ID, CreatedAt: e.CreatedAt}
	}
	if err := tx.Table(r.table).Create(&records).Error; err != nil {
		return fmt.Errorf("insert %s: %w", r.table, err)
	}
	return nil
}
