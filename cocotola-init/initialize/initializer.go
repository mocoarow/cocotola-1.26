// Package initialize provides the bootstrap initialization logic for cocotola-init.
package initialize

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

const (
	organizationName = "cocotola"
	organizationID   = 2
	maxActiveUsers   = 100
	maxActiveGroups  = 100
)

// Initialize bootstraps the cocotola tenant organization, creates the first owner user,
// adds them to the active user list, sets up RBAC policies, and assigns the admin group.
func Initialize(ctx context.Context, db *gorm.DB, ownerLoginID string, ownerPassword string) error {
	logger := slog.Default()

	// 1. Find or create "cocotola" organization
	orgRepo := gateway.NewOrganizationRepository(db)
	org, err := findOrCreateOrganization(ctx, orgRepo, logger)
	if err != nil {
		return fmt.Errorf("find or create organization: %w", err)
	}

	// 2. Create or find first owner user
	appUserRepo := gateway.NewAppUserRepository(db)
	hasher := gateway.NewBcryptHasher()
	ownerUserID, err := findOrCreateOwner(ctx, appUserRepo, hasher, org.ID(), ownerLoginID, ownerPassword, logger)
	if err != nil {
		return fmt.Errorf("find or create owner: %w", err)
	}

	// 3. Update active user list
	activeUserListRepo := gateway.NewActiveUserListRepository(db)
	if err := addToActiveUserList(ctx, activeUserListRepo, org, ownerUserID, logger); err != nil {
		return fmt.Errorf("add to active user list: %w", err)
	}

	// 4. Setup RBAC policies
	rbacRepo, err := gateway.NewRBACRepository(db)
	if err != nil {
		return fmt.Errorf("new rbac repository: %w", err)
	}
	if err := setupRBACPolicies(ctx, rbacRepo, org.ID(), logger); err != nil {
		return fmt.Errorf("setup rbac policies: %w", err)
	}

	// 5. Assign admin group to first owner
	if err := assignAdminGroup(ctx, rbacRepo, org.ID(), ownerUserID, logger); err != nil {
		return fmt.Errorf("assign admin group: %w", err)
	}

	// 6. Create or find guest user
	guestUserID, err := findOrCreateGuest(ctx, appUserRepo, org.ID(), logger)
	if err != nil {
		return fmt.Errorf("find or create guest: %w", err)
	}

	// 7. Add guest user to active user list
	if err := addToActiveUserList(ctx, activeUserListRepo, org, guestUserID, logger); err != nil {
		return fmt.Errorf("add guest to active user list: %w", err)
	}

	// 8. Create or find SystemOwner user
	systemOwnerID, err := findOrCreateSystemOwner(ctx, appUserRepo, hasher, org.ID(), logger)
	if err != nil {
		return fmt.Errorf("find or create system owner: %w", err)
	}

	// 9. Setup SystemOwner RBAC policies
	if err := setupSystemOwnerRBACPolicies(ctx, rbacRepo, org.ID(), logger); err != nil {
		return fmt.Errorf("setup system owner rbac policies: %w", err)
	}

	// 10. Assign system_owner group to SystemOwner user
	if err := assignSystemOwnerGroup(ctx, rbacRepo, org.ID(), systemOwnerID, logger); err != nil {
		return fmt.Errorf("assign system owner group: %w", err)
	}

	logger.InfoContext(ctx, "initialization completed successfully",
		slog.Int("organization_id", org.ID()),
		slog.Int("owner_user_id", ownerUserID),
		slog.Int("guest_user_id", guestUserID),
		slog.Int("system_owner_user_id", systemOwnerID),
	)
	return nil
}

func findOrCreateOrganization(ctx context.Context, repo *gateway.OrganizationRepository, logger *slog.Logger) (*domain.Organization, error) {
	org, err := repo.FindByName(ctx, organizationName)
	if err == nil {
		logger.InfoContext(ctx, "organization already exists",
			slog.Int("id", org.ID()),
			slog.String("name", org.Name()),
		)
		return org, nil
	}
	if !errors.Is(err, domain.ErrOrganizationNotFound) {
		return nil, fmt.Errorf("find organization by name: %w", err)
	}

	org, err = domain.NewOrganization(organizationID, organizationName, maxActiveUsers, maxActiveGroups)
	if err != nil {
		return nil, fmt.Errorf("new organization: %w", err)
	}

	if err := repo.Save(ctx, org); err != nil {
		return nil, fmt.Errorf("save organization: %w", err)
	}

	logger.InfoContext(ctx, "organization created",
		slog.Int("id", org.ID()),
		slog.String("name", org.Name()),
	)
	return org, nil
}

func findOrCreateOwner(ctx context.Context, repo *gateway.AppUserRepository, hasher domain.PasswordHasher, orgID int, loginID string, rawPassword string, logger *slog.Logger) (int, error) {
	user, err := repo.FindByLoginID(ctx, orgID, loginID)
	if err == nil {
		logger.InfoContext(ctx, "owner user already exists",
			slog.Int("user_id", user.ID()),
			slog.String("login_id", string(user.LoginID())),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return 0, fmt.Errorf("find user by login id: %w", err)
	}

	hashedPassword, err := domain.HashPassword(rawPassword, hasher)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	userID, err := repo.Create(ctx, orgID, loginID, hashedPassword)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	logger.InfoContext(ctx, "owner user created",
		slog.Int("user_id", userID),
		slog.String("login_id", loginID),
	)
	return userID, nil
}

func findOrCreateGuest(ctx context.Context, repo *gateway.AppUserRepository, orgID int, logger *slog.Logger) (int, error) {
	guestLoginID := domain.NewGuestLoginID(organizationName)

	user, err := repo.FindByLoginID(ctx, orgID, guestLoginID)
	if err == nil {
		logger.InfoContext(ctx, "guest user already exists",
			slog.Int("user_id", user.ID()),
			slog.String("login_id", string(user.LoginID())),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return 0, fmt.Errorf("find guest by login id: %w", err)
	}

	userID, err := repo.Create(ctx, orgID, guestLoginID, "")
	if err != nil {
		return 0, fmt.Errorf("create guest user: %w", err)
	}

	logger.InfoContext(ctx, "guest user created",
		slog.Int("user_id", userID),
		slog.String("login_id", guestLoginID),
	)
	return userID, nil
}

const systemOwnerLoginID = "__system_owner"

func findOrCreateSystemOwner(ctx context.Context, repo *gateway.AppUserRepository, hasher domain.PasswordHasher, orgID int, logger *slog.Logger) (int, error) {
	user, err := repo.FindByLoginID(ctx, orgID, systemOwnerLoginID)
	if err == nil {
		logger.InfoContext(ctx, "system owner user already exists",
			slog.Int("user_id", user.ID()),
			slog.String("login_id", string(user.LoginID())),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return 0, fmt.Errorf("find system owner by login id: %w", err)
	}

	// SystemOwner uses a random long password since it cannot login.
	dummyPassword := "system_owner_no_login_00000000"
	hashedPassword, err := domain.HashPassword(dummyPassword, hasher)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	userID, err := repo.Create(ctx, orgID, systemOwnerLoginID, hashedPassword)
	if err != nil {
		return 0, fmt.Errorf("create system owner: %w", err)
	}

	logger.InfoContext(ctx, "system owner user created",
		slog.Int("user_id", userID),
		slog.String("login_id", systemOwnerLoginID),
	)
	return userID, nil
}

func addToActiveUserList(ctx context.Context, repo *gateway.ActiveUserListRepository, org *domain.Organization, userID int, logger *slog.Logger) error {
	list, err := repo.FindByOrganizationID(ctx, org.ID())
	if err != nil {
		return fmt.Errorf("find active user list: %w", err)
	}

	if err := list.Add(userID, org.MaxActiveUsers()); err != nil {
		if errors.Is(err, domain.ErrDuplicateEntry) {
			logger.InfoContext(ctx, "user already in active user list", slog.Int("user_id", userID))
			return nil
		}
		return fmt.Errorf("add user to active list: %w", err)
	}

	if err := repo.Save(ctx, list); err != nil {
		return fmt.Errorf("save active user list: %w", err)
	}

	logger.InfoContext(ctx, "user added to active user list", slog.Int("user_id", userID))
	return nil
}

func setupRBACPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID int, logger *slog.Logger) error {
	adminGroup, err := domain.NewRBACGroup("admin")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	actions := []domain.RBACAction{
		domain.ActionCreateUser(),
		domain.ActionViewUser(),
		domain.ActionDisableUser(),
		domain.ActionChangePassword(),
		domain.ActionCreateGroup(),
		domain.ActionViewGroup(),
		domain.ActionDisableGroup(),
		domain.ActionAddUserToGroup(),
		domain.ActionRemoveUserFromGroup(),
	}

	for _, action := range actions {
		if err := repo.AddPolicy(ctx, orgID, adminGroup, action, domain.ResourceAny(), domain.EffectAllow()); err != nil {
			return fmt.Errorf("add policy %s: %w", action.Value(), err)
		}
	}

	logger.InfoContext(ctx, "RBAC policies created",
		slog.Int("organization_id", orgID),
		slog.Int("policy_count", len(actions)),
	)
	return nil
}

func setupSystemOwnerRBACPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID int, logger *slog.Logger) error {
	systemOwnerGroup, err := domain.NewRBACGroup("system_owner")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	actions := []domain.RBACAction{
		domain.ActionCreateUser(),
		domain.ActionViewUser(),
		domain.ActionDisableUser(),
		domain.ActionChangePassword(),
		domain.ActionCreateGroup(),
		domain.ActionViewGroup(),
		domain.ActionDisableGroup(),
		domain.ActionAddUserToGroup(),
		domain.ActionRemoveUserFromGroup(),
		domain.ActionCreateOrganization(),
	}

	for _, action := range actions {
		if err := repo.AddPolicy(ctx, orgID, systemOwnerGroup, action, domain.ResourceAny(), domain.EffectAllow()); err != nil {
			return fmt.Errorf("add system owner policy %s: %w", action.Value(), err)
		}
	}

	logger.InfoContext(ctx, "system owner RBAC policies created",
		slog.Int("organization_id", orgID),
		slog.Int("policy_count", len(actions)),
	)
	return nil
}

func assignAdminGroup(ctx context.Context, repo *gateway.RBACRepository, orgID int, userID int, logger *slog.Logger) error {
	adminGroup, err := domain.NewRBACGroup("admin")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	if err := repo.AssignGroupToUser(ctx, orgID, userID, adminGroup); err != nil {
		return fmt.Errorf("assign group to user: %w", err)
	}

	logger.InfoContext(ctx, "admin group assigned to user",
		slog.Int("organization_id", orgID),
		slog.Int("user_id", userID),
	)
	return nil
}

func assignSystemOwnerGroup(ctx context.Context, repo *gateway.RBACRepository, orgID int, userID int, logger *slog.Logger) error {
	systemOwnerGroup, err := domain.NewRBACGroup("system_owner")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	if err := repo.AssignGroupToUser(ctx, orgID, userID, systemOwnerGroup); err != nil {
		return fmt.Errorf("assign system owner group to user: %w", err)
	}

	logger.InfoContext(ctx, "system owner group assigned to user",
		slog.Int("organization_id", orgID),
		slog.Int("user_id", userID),
	)
	return nil
}
