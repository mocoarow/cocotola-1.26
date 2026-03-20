package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type userRecord struct {
	ID               int    `gorm:"column:id;primaryKey"`
	OrganizationID   int    `gorm:"column:organization_id"`
	LoginID          string `gorm:"column:login_id"`
	HashedPassword   string `gorm:"column:hashed_password"`
	Enabled          bool   `gorm:"column:enabled"`
	OrganizationName string `gorm:"column:organization_name"`
}

// UserAuthenticator verifies user credentials against the database.
type UserAuthenticator struct {
	db     *gorm.DB
	hasher domain.PasswordHasher
}

// NewUserAuthenticator returns a new UserAuthenticator.
func NewUserAuthenticator(db *gorm.DB, hasher domain.PasswordHasher) *UserAuthenticator {
	return &UserAuthenticator{db: db, hasher: hasher}
}

// Authenticate verifies the login credentials and returns user info.
func (a *UserAuthenticator) Authenticate(ctx context.Context, loginID string, password string, organizationName string) (*authservice.UserInfo, error) {
	var record userRecord
	err := a.db.WithContext(ctx).
		Table("app_user").
		Select("app_user.id, app_user.organization_id, app_user.login_id, app_user.hashed_password, app_user.enabled, organization.name as organization_name").
		Joins("JOIN organization ON app_user.organization_id = organization.id").
		Where("app_user.login_id = ? AND organization.name = ?", loginID, organizationName).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUnauthenticated
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !record.Enabled {
		return nil, domain.ErrUnauthenticated
	}

	user := domain.ReconstructAppUser(record.ID, record.OrganizationID, domain.LoginID(record.LoginID), record.HashedPassword, record.Enabled)
	if err := user.VerifyPassword(password, a.hasher); err != nil {
		return nil, domain.ErrUnauthenticated
	}

	userInfo, err := authservice.NewUserInfo(record.ID, record.LoginID, record.OrganizationName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("create user info: %w", err)
	}

	return userInfo, nil
}
