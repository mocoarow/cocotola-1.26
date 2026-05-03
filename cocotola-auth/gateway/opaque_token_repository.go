package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

// --- refresh token ---

type refreshTokenRecord struct {
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

func (refreshTokenRecord) TableName() string { return "refresh_token" }

func toRefreshTokenDomain(r *refreshTokenRecord) *domaintoken.RefreshToken {
	return domaintoken.ReconstructRefreshToken(r.ID, domain.MustParseAppUserID(r.UserID), domain.LoginID(r.LoginID), r.OrganizationName, domain.TokenHash(r.TokenHash), r.CreatedAt, r.ExpiresAt, r.RevokedAt)
}

// RefreshTokenRepository implements refresh token persistence using GORM.
type RefreshTokenRepository struct{ db *gorm.DB }

// NewRefreshTokenRepository returns a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Save persists a refresh token record (upsert: insert or update).
func (r *RefreshTokenRepository) Save(ctx context.Context, token *domaintoken.RefreshToken) error {
	record := refreshTokenRecord{
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
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

// FindByTokenHash looks up a refresh token by its SHA256 hash.
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*domaintoken.RefreshToken, error) {
	record, err := findRecordByHash[refreshTokenRecord](ctx, r.db, hash, "refresh token")
	if err != nil {
		return nil, fmt.Errorf("find refresh token by hash: %w", err)
	}
	return toRefreshTokenDomain(record), nil
}

// --- session token ---

type sessionTokenRecord struct {
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

func (sessionTokenRecord) TableName() string { return "session_token" }

func toSessionTokenDomain(r *sessionTokenRecord) *domaintoken.SessionToken {
	return domaintoken.ReconstructSessionToken(r.ID, domain.MustParseAppUserID(r.UserID), domain.LoginID(r.LoginID), r.OrganizationName, domain.TokenHash(r.TokenHash), r.CreatedAt, r.ExpiresAt, r.RevokedAt)
}

// SessionTokenRepository implements session token persistence using GORM.
type SessionTokenRepository struct{ db *gorm.DB }

// NewSessionTokenRepository returns a new SessionTokenRepository.
func NewSessionTokenRepository(db *gorm.DB) *SessionTokenRepository {
	return &SessionTokenRepository{db: db}
}

// Save persists a session token record (upsert: insert or update).
func (r *SessionTokenRepository) Save(ctx context.Context, token *domaintoken.SessionToken) error {
	record := sessionTokenRecord{
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
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save session token: %w", err)
	}
	return nil
}

// FindByTokenHash looks up a session token by its SHA256 hash.
func (r *SessionTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*domaintoken.SessionToken, error) {
	record, err := findRecordByHash[sessionTokenRecord](ctx, r.db, hash, "session token")
	if err != nil {
		return nil, fmt.Errorf("find session token by hash: %w", err)
	}
	return toSessionTokenDomain(record), nil
}
