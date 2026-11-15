package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

const (
	refreshTokenTable = "refresh_token"
	sessionTokenTable = "session_token"
)

type opaqueTokenRecord struct {
	ID               string     `gorm:"column:id;primaryKey"`
	Version          int        `gorm:"column:version"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	UserID           string     `gorm:"column:user_id"`
	LoginID          string     `gorm:"column:login_id"`
	OrganizationName string     `gorm:"column:organization_name"`
	TokenHash        string     `gorm:"column:token_hash"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
}

type opaqueToken interface {
	ID() string
	UserID() domain.AppUserID
	LoginID() domain.LoginID
	OrganizationName() string
	TokenHash() domain.TokenHash
	CreatedAt() time.Time
	ExpiresAt() time.Time
	RevokedAt() *time.Time
}

type reconstructOpaqueTokenFunc[T any] func(id string, userID domain.AppUserID, loginID domain.LoginID, organizationName string, tokenHash domain.TokenHash, createdAt time.Time, expiresAt time.Time, revokedAt *time.Time) *T

func saveOpaqueToken(ctx context.Context, db *gorm.DB, table string, token opaqueToken) error {
	record := opaqueTokenRecord{
		ID:               token.ID(),
		Version:          1,
		CreatedAt:        token.CreatedAt(),
		UpdatedAt:        time.Now(),
		UserID:           token.UserID().String(),
		LoginID:          string(token.LoginID()),
		OrganizationName: token.OrganizationName(),
		TokenHash:        string(token.TokenHash()),
		ExpiresAt:        token.ExpiresAt(),
		RevokedAt:        token.RevokedAt(),
	}
	if err := db.WithContext(ctx).Table(table).Save(&record).Error; err != nil {
		return fmt.Errorf("save %s: %w", table, err)
	}
	return nil
}

// findOpaqueTokenByHash looks up a token in the given table and returns ErrTokenNotFound when none matches.
func findOpaqueTokenByHash[T any](ctx context.Context, db *gorm.DB, table string, hash string, reconstruct reconstructOpaqueTokenFunc[T]) (*T, error) {
	var record opaqueTokenRecord
	if err := db.WithContext(ctx).Table(table).Where("token_hash = ?", hash).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find %s by hash: %w", table, err)
	}
	userID, err := domain.ParseAppUserID(record.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id %q in %s: %w", record.UserID, table, err)
	}
	return reconstruct(record.ID, userID, domain.LoginID(record.LoginID), record.OrganizationName, domain.TokenHash(record.TokenHash), record.CreatedAt, record.ExpiresAt, record.RevokedAt), nil
}

// RefreshTokenRepository implements refresh token persistence using GORM.
type RefreshTokenRepository struct{ db *gorm.DB }

// NewRefreshTokenRepository returns a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Save persists a refresh token record (upsert: insert or update).
func (r *RefreshTokenRepository) Save(ctx context.Context, token *domaintoken.RefreshToken) error {
	return saveOpaqueToken(ctx, r.db, refreshTokenTable, token)
}

// FindByTokenHash looks up a refresh token by its SHA256 hash.
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*domaintoken.RefreshToken, error) {
	return findOpaqueTokenByHash(ctx, r.db, refreshTokenTable, hash, domaintoken.ReconstructRefreshToken)
}

// SessionTokenRepository implements session token persistence using GORM.
type SessionTokenRepository struct{ db *gorm.DB }

// NewSessionTokenRepository returns a new SessionTokenRepository.
func NewSessionTokenRepository(db *gorm.DB) *SessionTokenRepository {
	return &SessionTokenRepository{db: db}
}

// Save persists a session token record (upsert: insert or update).
func (r *SessionTokenRepository) Save(ctx context.Context, token *domaintoken.SessionToken) error {
	return saveOpaqueToken(ctx, r.db, sessionTokenTable, token)
}

// FindByTokenHash looks up a session token by its SHA256 hash.
func (r *SessionTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*domaintoken.SessionToken, error) {
	return findOpaqueTokenByHash(ctx, r.db, sessionTokenTable, hash, domaintoken.ReconstructSessionToken)
}
