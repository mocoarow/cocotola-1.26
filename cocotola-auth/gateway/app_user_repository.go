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

// initialAppUserVersion is the row version assigned to a brand-new AppUser on
// its first INSERT. Persisting at version 1 (rather than 0) keeps the in-memory
// aggregate and the stored row in sync after IncrementVersion, so a subsequent
// Save on the same aggregate can CAS against the expected version.
const initialAppUserVersion = 1

type appUserRecord struct {
	ID                            int       `gorm:"column:id;primaryKey"`
	Version                       int       `gorm:"column:version"`
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
	var provider string
	if r.Provider != nil {
		provider = *r.Provider
	}
	var providerID string
	if r.ProviderID != nil {
		providerID = *r.ProviderID
	}
	return domainuser.
		ReconstructAppUser(r.ID, r.OrganizationID, domain.LoginID(r.LoginID), hashedPw, provider, providerID, r.Enabled).
		WithVersion(r.Version)
}

func toAppUserRecord(user *domainuser.AppUser) appUserRecord {
	var hashedPw *string
	if hp := user.HashedPassword(); hp != "" {
		hashedPw = &hp
	}
	var provider *string
	if p := user.Provider(); p != "" {
		provider = &p
	}
	var providerID *string
	if pid := user.ProviderID(); pid != "" {
		providerID = &pid
	}
	return appUserRecord{
		ID:                            user.ID(),
		Version:                       user.Version(),
		CreatedAt:                     time.Time{},
		UpdatedAt:                     time.Time{},
		CreatedBy:                     0,
		UpdatedBy:                     0,
		OrganizationID:                user.OrganizationID(),
		LoginID:                       string(user.LoginID()),
		HashedPassword:                hashedPw,
		Username:                      nil,
		Provider:                      provider,
		ProviderID:                    providerID,
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

// NextID reserves and returns the next app user ID from the database sequence.
// Allowing aggregates to receive their identity at construction time keeps the
// repository as a pure collection (Save persists, no factory methods needed).
func (r *AppUserRepository) NextID(ctx context.Context) (int, error) {
	var id int
	if err := r.db.WithContext(ctx).
		Raw("SELECT nextval(pg_get_serial_sequence('app_user', 'id'))").
		Scan(&id).Error; err != nil {
		return 0, fmt.Errorf("next app user id: %w", err)
	}
	return id, nil
}

// Save persists an app user aggregate as a whole. New aggregates (version 0)
// are inserted; loaded aggregates (version > 0) are updated via a compare-and-swap
// on the version column so concurrent updates cannot silently clobber each other.
// On success, the aggregate's in-memory version is bumped.
//
// Returns domain.ErrAppUserConcurrentModification when the CAS finds no matching
// row, indicating the aggregate was modified by another transaction after it was
// loaded and must be reloaded before retrying.
//
// Columns not tracked by the AppUser aggregate (username, encrypted tokens) are
// left untouched.
func (r *AppUserRepository) Save(ctx context.Context, user *domainuser.AppUser) error {
	record := toAppUserRecord(user)
	if user.Version() == 0 {
		record.Version = initialAppUserVersion
		if err := r.db.WithContext(ctx).
			Omit("username", "encrypted_provider_access_token", "encrypted_provider_refresh_token").
			Create(&record).Error; err != nil {
			return fmt.Errorf("insert app user: %w", err)
		}
		user.IncrementVersion()
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&record).
		Where("id = ? AND version = ?", record.ID, record.Version).
		Updates(map[string]any{
			"organization_id": record.OrganizationID,
			"login_id":        record.LoginID,
			"hashed_password": record.HashedPassword,
			"provider":        record.Provider,
			"provider_id":     record.ProviderID,
			"enabled":         record.Enabled,
			"version":         gorm.Expr("app_user.version + 1"),
		})
	if result.Error != nil {
		return fmt.Errorf("update app user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrAppUserConcurrentModification
	}
	user.IncrementVersion()
	return nil
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
func (r *AppUserRepository) FindByLoginID(ctx context.Context, organizationID int, loginID domain.LoginID) (*domainuser.AppUser, error) {
	var record appUserRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND login_id = ?", organizationID, string(loginID)).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserNotFound
		}
		return nil, fmt.Errorf("find app user by login id: %w", err)
	}
	return toAppUserDomain(&record), nil
}

// FindByProviderID looks up an app user by organization, provider, and provider ID.
func (r *AppUserRepository) FindByProviderID(ctx context.Context, organizationID int, provider string, providerID string) (*domainuser.AppUser, error) {
	var record appUserRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND provider = ? AND provider_id = ?", organizationID, provider, providerID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAppUserNotFound
		}
		return nil, fmt.Errorf("find app user by provider id: %w", err)
	}
	return toAppUserDomain(&record), nil
}
