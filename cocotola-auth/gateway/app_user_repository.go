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
	ID                            string    `gorm:"column:id;primaryKey"`
	Version                       int       `gorm:"column:version"`
	CreatedAt                     time.Time `gorm:"column:created_at;->"`
	UpdatedAt                     time.Time `gorm:"column:updated_at;->"`
	CreatedBy                     string    `gorm:"column:created_by;<-:create"`
	UpdatedBy                     string    `gorm:"column:updated_by"`
	OrganizationID                string    `gorm:"column:organization_id"`
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
	u := domainuser.ReconstructAppUser(domain.MustParseAppUserID(r.ID), domain.MustParseOrganizationID(r.OrganizationID), domain.LoginID(r.LoginID), hashedPw, r.Enabled)
	u.SetVersion(r.Version)
	return u
}

func toAppUserRecord(user *domainuser.AppUser) appUserRecord {
	var hashedPw *string
	if hp := user.HashedPassword(); hp != "" {
		hashedPw = &hp
	}
	systemUserID := domain.SystemAppUserID().String()
	return appUserRecord{
		ID:                            user.ID().String(),
		Version:                       user.Version(),
		CreatedAt:                     time.Time{},
		UpdatedAt:                     time.Time{},
		CreatedBy:                     systemUserID,
		UpdatedBy:                     systemUserID,
		OrganizationID:                user.OrganizationID().String(),
		LoginID:                       string(user.LoginID()),
		HashedPassword:                hashedPw,
		Username:                      nil,
		Provider:                      nil,
		ProviderID:                    nil,
		EncryptedProviderAccessToken:  nil,
		EncryptedProviderRefreshToken: nil,
		Enabled:                       user.Enabled(),
	}
}

// AppUserRepository implements app user persistence using GORM.
type AppUserRepository struct {
	db *gorm.DB
}

// NewAppUserRepository returns a new AppUserRepository.
func NewAppUserRepository(db *gorm.DB) *AppUserRepository {
	return &AppUserRepository{db: db}
}

// Save persists an app user aggregate. New aggregates (version 0) are inserted;
// loaded aggregates (version > 0) are updated via CAS on the version column.
// The repository updates the aggregate's version after a successful persist so
// the caller does not need to manage versioning.
func (r *AppUserRepository) Save(ctx context.Context, user *domainuser.AppUser) error {
	record := toAppUserRecord(user)
	nextVersion := user.Version() + 1
	if user.Version() == 0 {
		record.Version = nextVersion
		if err := r.db.WithContext(ctx).
			Omit("username", "encrypted_provider_access_token", "encrypted_provider_refresh_token").
			Create(&record).Error; err != nil {
			return fmt.Errorf("insert app user: %w", err)
		}
		user.SetVersion(nextVersion)
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&record).
		Where("id = ? AND version = ?", record.ID, user.Version()).
		Updates(map[string]any{
			"organization_id": record.OrganizationID,
			"login_id":        record.LoginID,
			"hashed_password": record.HashedPassword,
			"enabled":         record.Enabled,
			"version":         nextVersion,
		})
	if result.Error != nil {
		return fmt.Errorf("update app user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrAppUserConcurrentModification
	}
	user.SetVersion(nextVersion)
	return nil
}

// FindByID looks up an app user by its ID.
func (r *AppUserRepository) FindByID(ctx context.Context, id domain.AppUserID) (*domainuser.AppUser, error) {
	var record appUserRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserNotFound
		}
		return nil, fmt.Errorf("find app user by id: %w", err)
	}
	return toAppUserDomain(&record), nil
}

// FindByLoginID looks up an app user by organization ID and login ID.
func (r *AppUserRepository) FindByLoginID(ctx context.Context, organizationID domain.OrganizationID, loginID domain.LoginID) (*domainuser.AppUser, error) {
	var record appUserRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND login_id = ?", organizationID.String(), string(loginID)).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserNotFound
		}
		return nil, fmt.Errorf("find app user by login id: %w", err)
	}
	return toAppUserDomain(&record), nil
}
