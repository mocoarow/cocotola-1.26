package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
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

func toAppUserDomain(r *appUserRecord) *domainuser.AppUser {
	var hashedPw string
	if r.HashedPassword != nil {
		hashedPw = *r.HashedPassword
	}
	return domainuser.ReconstructAppUser(r.ID, r.OrganizationID, domain.LoginID(r.LoginID), hashedPw, r.Enabled)
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
func (r *AppUserRepository) Save(ctx context.Context, user *domainuser.AppUser) error {
	hp := user.HashedPassword()
	var hashedPw *string
	if hp != "" {
		hashedPw = &hp
	}
	record := appUserRecord{
		ID:                            user.ID(),
		Version:                       0,
		CreatedAt:                     time.Time{},
		UpdatedAt:                     time.Time{},
		CreatedBy:                     0,
		UpdatedBy:                     0,
		OrganizationID:                user.OrganizationID(),
		LoginID:                       string(user.LoginID()),
		HashedPassword:                hashedPw,
		Username:                      nil,
		Provider:                      nil,
		ProviderID:                    nil,
		EncryptedProviderAccessToken:  nil,
		EncryptedProviderRefreshToken: nil,
		Enabled:                       user.Enabled(),
	}
	if err := r.db.WithContext(ctx).
		Omit("username", "provider", "provider_id", "encrypted_provider_access_token", "encrypted_provider_refresh_token").
		Save(&record).Error; err != nil {
		return fmt.Errorf("save app user: %w", err)
	}
	return nil
}

// Create inserts a new app user record and returns the auto-generated ID.
func (r *AppUserRepository) Create(ctx context.Context, organizationID int, loginID string, hashedPassword string) (int, error) {
	record := appUserRecord{
		ID:                            0,
		Version:                       0,
		CreatedAt:                     time.Time{},
		UpdatedAt:                     time.Time{},
		CreatedBy:                     0,
		UpdatedBy:                     0,
		OrganizationID:                organizationID,
		LoginID:                       loginID,
		HashedPassword:                &hashedPassword,
		Username:                      nil,
		Provider:                      nil,
		ProviderID:                    nil,
		EncryptedProviderAccessToken:  nil,
		EncryptedProviderRefreshToken: nil,
		Enabled:                       true,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return 0, fmt.Errorf("create app user: %w", err)
	}
	return record.ID, nil
}

// FindByID looks up an app user by its ID.
func (r *AppUserRepository) FindByID(ctx context.Context, id int) (*domainuser.AppUser, error) {
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
func (r *AppUserRepository) FindByLoginID(ctx context.Context, organizationID int, loginID string) (*domainuser.AppUser, error) {
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
