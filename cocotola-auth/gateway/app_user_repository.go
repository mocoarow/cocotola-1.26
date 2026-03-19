package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type appUserRecord struct {
	ID                            int       `gorm:"column:id;primaryKey"`
	Version                       int       `gorm:"column:version;->"`
	CreatedAt                     time.Time `gorm:"column:created_at;->"`
	UpdatedAt                     time.Time `gorm:"column:updated_at;->"`
	CreatedBy                     int       `gorm:"column:created_by;<-:create"`
	UpdatedBy                     int       `gorm:"column:updated_by"`
	OrganizationID                int       `gorm:"column:organization_id"`
	LoginID                       string    `gorm:"column:login_id"`
	HashedPassword                *string   `gorm:"column:hashed_password"`
	Username                      *string   `gorm:"column:username"`
	Provider                      *string   `gorm:"column:provider"`
	ProviderID                    *string   `gorm:"column:provider_id"`
	EncryptedProviderAccessToken  *string   `gorm:"column:encrypted_provider_access_token"`
	EncryptedProviderRefreshToken *string   `gorm:"column:encrypted_provider_refresh_token"`
	Enabled                       bool      `gorm:"column:enabled"`
}

func (appUserRecord) TableName() string {
	return "app_user"
}

func toAppUserDomain(r *appUserRecord) *domain.AppUser {
	return domain.ReconstructAppUser(r.ID, r.OrganizationID, domain.LoginID(r.LoginID), r.Enabled)
}

// AppUserRepository implements app user persistence using MySQL via GORM.
type AppUserRepository struct {
	db *gorm.DB
}

// NewAppUserRepository returns a new AppUserRepository.
func NewAppUserRepository(db *gorm.DB) *AppUserRepository {
	return &AppUserRepository{db: db}
}

// Save persists an app user record (upsert: insert or update).
func (r *AppUserRepository) Save(ctx context.Context, user *domain.AppUser) error {
	record := appUserRecord{
		ID:             user.ID(),
		OrganizationID: user.OrganizationID(),
		LoginID:        string(user.LoginID()),
		Enabled:        user.Enabled(),
	}
	if err := r.db.WithContext(ctx).
		Omit("hashed_password", "username", "provider", "provider_id", "encrypted_provider_access_token", "encrypted_provider_refresh_token").
		Save(&record).Error; err != nil {
		return fmt.Errorf("save app user: %w", err)
	}
	return nil
}

// FindByID looks up an app user by its ID.
func (r *AppUserRepository) FindByID(ctx context.Context, id int) (*domain.AppUser, error) {
	var record appUserRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserNotFound
		}
		return nil, fmt.Errorf("find app user by id: %w", err)
	}
	return toAppUserDomain(&record), nil
}

// FindByLoginID looks up an app user by organization ID and login ID.
func (r *AppUserRepository) FindByLoginID(ctx context.Context, organizationID int, loginID string) (*domain.AppUser, error) {
	var record appUserRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND login_id = ?", organizationID, loginID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserNotFound
		}
		return nil, fmt.Errorf("find app user by login id: %w", err)
	}
	return toAppUserDomain(&record), nil
}
