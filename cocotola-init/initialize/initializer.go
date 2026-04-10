// Package initialize provides the bootstrap initialization logic for cocotola-init.
package initialize

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

const (
	organizationName = "cocotola"
	// CocotolaOrganizationIDString is the well-known UUID of the "cocotola" tenant
	// organization bootstrapped by cocotola-init. It must remain stable across
	// deployments so that rows referencing it stay valid.
	CocotolaOrganizationIDString = "00000000-0000-7000-8000-000000000100"
	maxActiveUsers               = 100
	maxActiveGroups              = 100
)

func cocotolaOrganizationID() domain.OrganizationID {
	return domain.MustParseOrganizationID(CocotolaOrganizationIDString)
}

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

	// 8. Setup system owner and public space
	if err := setupSystemOwnerAndSpace(ctx, db, appUserRepo, hasher, rbacRepo, org.ID(), logger); err != nil {
		return fmt.Errorf("setup system owner and space: %w", err)
	}

	logger.InfoContext(ctx, "initialization completed successfully",
		slog.String("organization_id", org.ID().String()),
		slog.String("owner_user_id", ownerUserID.String()),
		slog.String("guest_user_id", guestUserID.String()),
	)
	return nil
}

func setupSystemOwnerAndSpace(ctx context.Context, db *gorm.DB, appUserRepo *gateway.AppUserRepository, hasher domainuser.PasswordHasher, rbacRepo *gateway.RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	systemOwnerID, err := findOrCreateSystemOwner(ctx, appUserRepo, hasher, orgID, logger)
	if err != nil {
		return fmt.Errorf("find or create system owner: %w", err)
	}

	if err := setupSystemOwnerRBACPolicies(ctx, rbacRepo, orgID, logger); err != nil {
		return fmt.Errorf("setup system owner rbac policies: %w", err)
	}

	if err := assignSystemOwnerGroup(ctx, rbacRepo, orgID, systemOwnerID, logger); err != nil {
		return fmt.Errorf("assign system owner group: %w", err)
	}

	spaceRepo := gateway.NewSpaceRepository(db)
	if err := findOrCreatePublicSpace(ctx, spaceRepo, orgID, systemOwnerID, logger); err != nil {
		return fmt.Errorf("find or create public space: %w", err)
	}

	return nil
}

func findOrCreateOrganization(ctx context.Context, repo *gateway.OrganizationRepository, logger *slog.Logger) (*domain.Organization, error) {
	org, err := repo.FindByName(ctx, organizationName)
	if err == nil {
		logger.InfoContext(ctx, "organization already exists",
			slog.String("id", org.ID().String()),
			slog.String("name", org.Name()),
		)
		return org, nil
	}
	if !errors.Is(err, domain.ErrOrganizationNotFound) {
		return nil, fmt.Errorf("find organization by name: %w", err)
	}

	org, err = domain.NewOrganization(cocotolaOrganizationID(), organizationName, maxActiveUsers, maxActiveGroups)
	if err != nil {
		return nil, fmt.Errorf("new organization: %w", err)
	}

	if err := repo.Save(ctx, org); err != nil {
		return nil, fmt.Errorf("save organization: %w", err)
	}

	logger.InfoContext(ctx, "organization created",
		slog.String("id", org.ID().String()),
		slog.String("name", org.Name()),
	)
	return org, nil
}

func findOrCreateOwner(ctx context.Context, repo *gateway.AppUserRepository, hasher domainuser.PasswordHasher, orgID domain.OrganizationID, loginID string, rawPassword string, logger *slog.Logger) (domain.AppUserID, error) {
	user, err := repo.FindByLoginID(ctx, orgID, domain.LoginID(loginID))
	if err == nil {
		logger.InfoContext(ctx, "owner user already exists",
			slog.String("user_id", user.ID().String()),
			slog.String("login_id", string(user.LoginID())),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return domain.AppUserID{}, fmt.Errorf("find user by login id: %w", err)
	}

	hashedPassword, err := domainuser.HashPassword(rawPassword, hasher)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("hash password: %w", err)
	}

	user, err = domainuser.Provision(ctx, repo, orgID, domain.LoginID(loginID), hashedPassword, "", "", true)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("provision owner user: %w", err)
	}

	logger.InfoContext(ctx, "owner user created",
		slog.String("user_id", user.ID().String()),
		slog.String("login_id", loginID),
	)
	return user.ID(), nil
}

func findOrCreateGuest(ctx context.Context, repo *gateway.AppUserRepository, orgID domain.OrganizationID, logger *slog.Logger) (domain.AppUserID, error) {
	guestLoginID := domainuser.NewGuestLoginID(organizationName)

	user, err := repo.FindByLoginID(ctx, orgID, domain.LoginID(guestLoginID))
	if err == nil {
		logger.InfoContext(ctx, "guest user already exists",
			slog.String("user_id", user.ID().String()),
			slog.String("login_id", string(user.LoginID())),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return domain.AppUserID{}, fmt.Errorf("find guest by login id: %w", err)
	}

	user, err = domainuser.Provision(ctx, repo, orgID, domain.LoginID(guestLoginID), "", "", "", true)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("provision guest user: %w", err)
	}

	logger.InfoContext(ctx, "guest user created",
		slog.String("user_id", user.ID().String()),
		slog.String("login_id", guestLoginID),
	)
	return user.ID(), nil
}

const systemOwnerLoginID = "__system_owner"

func findOrCreateSystemOwner(ctx context.Context, repo *gateway.AppUserRepository, hasher domainuser.PasswordHasher, orgID domain.OrganizationID, logger *slog.Logger) (domain.AppUserID, error) {
	user, err := repo.FindByLoginID(ctx, orgID, domain.LoginID(systemOwnerLoginID))
	if err == nil {
		logger.InfoContext(ctx, "system owner user already exists",
			slog.String("user_id", user.ID().String()),
			slog.String("login_id", string(user.LoginID())),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return domain.AppUserID{}, fmt.Errorf("find system owner by login id: %w", err)
	}

	// SystemOwner uses a random long password since it cannot login.
	dummyPassword := "system_owner_no_login_00000000"
	hashedPassword, err := domainuser.HashPassword(dummyPassword, hasher)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("hash password: %w", err)
	}

	user, err = domainuser.Provision(ctx, repo, orgID, domain.LoginID(systemOwnerLoginID), hashedPassword, "", "", true)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("provision system owner: %w", err)
	}

	logger.InfoContext(ctx, "system owner user created",
		slog.String("user_id", user.ID().String()),
		slog.String("login_id", systemOwnerLoginID),
	)
	return user.ID(), nil
}

func addToActiveUserList(ctx context.Context, repo *gateway.ActiveUserListRepository, org *domain.Organization, userID domain.AppUserID, logger *slog.Logger) error {
	list, err := repo.FindByOrganizationID(ctx, org.ID())
	if err != nil {
		return fmt.Errorf("find active user list: %w", err)
	}

	if err := list.Add(userID, org.MaxActiveUsers()); err != nil {
		if errors.Is(err, domain.ErrDuplicateEntry) {
			logger.InfoContext(ctx, "user already in active user list", slog.String("user_id", userID.String()))
			return nil
		}
		return fmt.Errorf("add user to active list: %w", err)
	}

	if err := repo.Save(ctx, list); err != nil {
		return fmt.Errorf("save active user list: %w", err)
	}

	logger.InfoContext(ctx, "user added to active user list", slog.String("user_id", userID.String()))
	return nil
}

func setupRBACPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	adminGroup, err := domainrbac.NewGroup("admin")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	actions := []domainrbac.Action{
		domainrbac.ActionCreateUser(),
		domainrbac.ActionListUsers(),
		domainrbac.ActionViewUser(),
		domainrbac.ActionDisableUser(),
		domainrbac.ActionChangePassword(),
		domainrbac.ActionCreateGroup(),
		domainrbac.ActionListGroups(),
		domainrbac.ActionViewGroup(),
		domainrbac.ActionDisableGroup(),
		domainrbac.ActionAddUserToGroup(),
		domainrbac.ActionRemoveUserFromGroup(),
		domainrbac.ActionCreateSpace(),
		domainrbac.ActionListSpaces(),
		domainrbac.ActionViewSpace(),
	}

	for _, action := range actions {
		if err := repo.AddPolicy(ctx, orgID, adminGroup, action, domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add policy %s: %w", action.Value(), err)
		}
	}

	logger.InfoContext(ctx, "RBAC policies created",
		slog.String("organization_id", orgID.String()),
		slog.Int("policy_count", len(actions)),
	)
	return nil
}

func setupSystemOwnerRBACPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	systemOwnerGroup, err := domainrbac.NewGroup("system_owner")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	actions := []domainrbac.Action{
		domainrbac.ActionCreateUser(),
		domainrbac.ActionListUsers(),
		domainrbac.ActionViewUser(),
		domainrbac.ActionDisableUser(),
		domainrbac.ActionChangePassword(),
		domainrbac.ActionCreateGroup(),
		domainrbac.ActionListGroups(),
		domainrbac.ActionViewGroup(),
		domainrbac.ActionDisableGroup(),
		domainrbac.ActionAddUserToGroup(),
		domainrbac.ActionRemoveUserFromGroup(),
		domainrbac.ActionCreateOrganization(),
		domainrbac.ActionCreateSpace(),
		domainrbac.ActionListSpaces(),
		domainrbac.ActionViewSpace(),
	}

	for _, action := range actions {
		if err := repo.AddPolicy(ctx, orgID, systemOwnerGroup, action, domainrbac.ResourceAny(), domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add system owner policy %s: %w", action.Value(), err)
		}
	}

	logger.InfoContext(ctx, "system owner RBAC policies created",
		slog.String("organization_id", orgID.String()),
		slog.Int("policy_count", len(actions)),
	)
	return nil
}

func assignAdminGroup(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, userID domain.AppUserID, logger *slog.Logger) error {
	adminGroup, err := domainrbac.NewGroup("admin")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	if err := repo.AssignGroupToUser(ctx, orgID, userID, adminGroup); err != nil {
		return fmt.Errorf("assign group to user: %w", err)
	}

	logger.InfoContext(ctx, "admin group assigned to user",
		slog.String("organization_id", orgID.String()),
		slog.String("user_id", userID.String()),
	)
	return nil
}

func findOrCreatePublicSpace(ctx context.Context, repo *gateway.SpaceRepository, orgID domain.OrganizationID, systemOwnerID domain.AppUserID, logger *slog.Logger) error {
	keyName := domainspace.PublicSpaceKeyName(organizationName)

	_, err := repo.FindByKeyName(ctx, orgID, keyName)
	if err == nil {
		logger.InfoContext(ctx, "public space already exists", slog.String("key_name", keyName))
		return nil
	}
	if !errors.Is(err, domain.ErrSpaceNotFound) {
		return fmt.Errorf("find public space by key name: %w", err)
	}

	spaceID, err := repo.Create(ctx, orgID, systemOwnerID, keyName, "Public", domainspace.TypePublic().Value(), systemOwnerID)
	if err != nil {
		return fmt.Errorf("create public space: %w", err)
	}

	logger.InfoContext(ctx, "public space created",
		slog.Int("space_id", spaceID),
		slog.String("key_name", keyName),
	)
	return nil
}

func assignSystemOwnerGroup(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, userID domain.AppUserID, logger *slog.Logger) error {
	systemOwnerGroup, err := domainrbac.NewGroup("system_owner")
	if err != nil {
		return fmt.Errorf("new rbac group: %w", err)
	}

	if err := repo.AssignGroupToUser(ctx, orgID, userID, systemOwnerGroup); err != nil {
		return fmt.Errorf("assign system owner group to user: %w", err)
	}

	logger.InfoContext(ctx, "system owner group assigned to user",
		slog.String("organization_id", orgID.String()),
		slog.String("user_id", userID.String()),
	)
	return nil
}
