package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type refreshTokenRecord struct {
	ID               string     `gorm:"column:id;primaryKey"`
	Version          int        `gorm:"column:version"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	UserID           int        `gorm:"column:user_id"`
	LoginID          string     `gorm:"column:login_id"`
	OrganizationName string     `gorm:"column:organization_name"`
	TokenHash        string     `gorm:"column:token_hash"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
}

func (refreshTokenRecord) TableName() string {
	return "refresh_token"
}

func toRefreshTokenDomain(r *refreshTokenRecord) *domain.RefreshToken {
	return domain.ReconstructRefreshToken(r.ID, r.UserID, domain.LoginID(r.LoginID), r.OrganizationName, domain.TokenHash(r.TokenHash), r.CreatedAt, r.ExpiresAt, r.RevokedAt)
}

// RefreshTokenRepository implements refresh token persistence using MySQL via GORM.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository returns a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Save persists a refresh token record (upsert: insert or update).
func (r *RefreshTokenRepository) Save(ctx context.Context, token *domain.RefreshToken) error {
	record := refreshTokenRecord{
		ID:               token.ID(),
		Version:          1,
		CreatedAt:        token.CreatedAt(),
		UpdatedAt:        time.Now(),
		UserID:           token.UserID(),
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
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var record refreshTokenRecord
	if err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find refresh token by hash: %w", err)
	}
	return toRefreshTokenDomain(&record), nil
}
