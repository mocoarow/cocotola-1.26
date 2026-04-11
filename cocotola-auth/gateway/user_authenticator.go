package gateway

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type userRecord struct {
	ID               string `gorm:"column:id;primaryKey"`
	OrganizationID   string `gorm:"column:organization_id"`
	LoginID          string `gorm:"column:login_id"`
	HashedPassword   string `gorm:"column:hashed_password"`
	Enabled          bool   `gorm:"column:enabled"`
	OrganizationName string `gorm:"column:organization_name"`
}

// UserAuthenticator verifies user credentials against the database.
type UserAuthenticator struct {
	db          *gorm.DB
	hasher      domainuser.PasswordHasher
	groupFinder domainrbac.GroupFinder
}

// NewUserAuthenticator returns a new UserAuthenticator.
func NewUserAuthenticator(db *gorm.DB, hasher domainuser.PasswordHasher, groupFinder domainrbac.GroupFinder) *UserAuthenticator {
	return &UserAuthenticator{db: db, hasher: hasher, groupFinder: groupFinder}
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

	userID := domain.MustParseAppUserID(record.ID)
	orgID := domain.MustParseOrganizationID(record.OrganizationID)
	user := domainuser.ReconstructAppUser(userID, orgID, domain.LoginID(record.LoginID), record.HashedPassword, record.Enabled)
	if err := user.VerifyPassword(password, a.hasher); err != nil {
		return nil, domain.ErrUnauthenticated
	}

	// Check login-denied groups via Casbin.
	groups, err := a.groupFinder.GetGroupsForUser(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("get groups for user: %w", err)
	}
	if slices.ContainsFunc(groups, domainrbac.IsLoginDenied) {
		return nil, domain.ErrUnauthenticated
	}

	userInfo, err := authservice.NewUserInfo(userID, record.LoginID, record.OrganizationName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("create user info: %w", err)
	}

	return userInfo, nil
}
