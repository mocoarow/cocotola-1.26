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

type accessTokenRecord struct {
	ID               string     `gorm:"column:id;primaryKey"`
	Version          int        `gorm:"column:version"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	RefreshTokenID   string     `gorm:"column:refresh_token_id"`
	UserID           int        `gorm:"column:user_id"`
	LoginID          string     `gorm:"column:login_id"`
	OrganizationName string     `gorm:"column:organization_name"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
}

func (accessTokenRecord) TableName() string {
	return "access_token"
}

func toAccessTokenDomain(r *accessTokenRecord) *domaintoken.AccessToken {
	return domaintoken.ReconstructAccessToken(r.ID, r.RefreshTokenID, r.UserID, domain.LoginID(r.LoginID), r.OrganizationName, r.CreatedAt, r.ExpiresAt, r.RevokedAt)
}

// AccessTokenRepository implements access token persistence using GORM.
type AccessTokenRepository struct {
	db *gorm.DB
}

// NewAccessTokenRepository returns a new AccessTokenRepository.
func NewAccessTokenRepository(db *gorm.DB) *AccessTokenRepository {
	return &AccessTokenRepository{db: db}
}

// Save persists an access token record (upsert: insert or update).
func (r *AccessTokenRepository) Save(ctx context.Context, token *domaintoken.AccessToken) error {
	record := accessTokenRecord{
		ID:               token.ID(),
		Version:          1,
		CreatedAt:        token.CreatedAt(),
		UpdatedAt:        time.Now(),
		RefreshTokenID:   token.RefreshTokenID(),
		UserID:           token.UserID(),
		LoginID:          string(token.LoginID()),
		OrganizationName: token.OrganizationName(),
		ExpiresAt:        token.ExpiresAt(),
		RevokedAt:        token.RevokedAt(),
	}
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save access token: %w", err)
	}
	return nil
}

// FindByID looks up an access token by its ID (= JWT JTI).
func (r *AccessTokenRepository) FindByID(ctx context.Context, id string) (*domaintoken.AccessToken, error) {
	var record accessTokenRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find access token by id: %w", err)
	}
	return toAccessTokenDomain(&record), nil
}

// FindByRefreshTokenID returns all access tokens that belong to the given refresh token.
func (r *AccessTokenRepository) FindByRefreshTokenID(ctx context.Context, refreshTokenID string) ([]domaintoken.AccessToken, error) {
	var records []accessTokenRecord
	if err := r.db.WithContext(ctx).
		Where("refresh_token_id = ?", refreshTokenID).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find access tokens by refresh token id: %w", err)
	}
	tokens := make([]domaintoken.AccessToken, len(records))
	for i := range records {
		tokens[i] = *toAccessTokenDomain(&records[i])
	}
	return tokens, nil
}
