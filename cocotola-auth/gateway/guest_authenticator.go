package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// GuestAuthenticator verifies guest credentials against the database.
type GuestAuthenticator struct {
	db *gorm.DB
}

// NewGuestAuthenticator returns a new GuestAuthenticator.
func NewGuestAuthenticator(db *gorm.DB) *GuestAuthenticator {
	return &GuestAuthenticator{db: db}
}

// Authenticate finds the guest user for the given organization and returns user info.
func (a *GuestAuthenticator) Authenticate(ctx context.Context, organizationName string) (*authservice.UserInfo, error) {
	loginID := domainuser.NewGuestLoginID(organizationName)

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
		return nil, fmt.Errorf("find guest user: %w", err)
	}

	if !record.Enabled {
		return nil, domain.ErrUnauthenticated
	}

	userInfo, err := authservice.NewUserInfo(record.ID, record.LoginID, record.OrganizationName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("create user info: %w", err)
	}

	return userInfo, nil
}
