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

// PublicWorkbookSeeder seeds public workbooks (and their questions) into
// cocotola-question via its internal HTTP API.
type PublicWorkbookSeeder interface {
	SeedPublicWorkbooks(ctx context.Context, organizationID, publicSpaceID string) error
}

const (
	organizationName = "cocotola"
	// CocotolaOrganizationIDString is the well-known UUID of the "cocotola" tenant
	// organization bootstrapped by cocotola-init. It must remain stable across
	// deployments so that rows referencing it stay valid.
	CocotolaOrganizationIDString = "00000000-0000-7000-8000-000000000100"
	maxActiveUsers               = 100
	maxActiveGroups              = 100
)

// CocotolaOrganizationID returns CocotolaOrganizationIDString as an ID.
func CocotolaOrganizationID() domain.OrganizationID {
	return domain.MustParseOrganizationID(CocotolaOrganizationIDString)
}

// Initialize bootstraps the cocotola tenant organization, creates the first owner user,
// adds them to the active user list, sets up RBAC policies, and assigns the admin group.
//
// seeder is invoked after the public space and SystemOwner policies are in place
// to populate the public workbooks. It must be non-nil.
func Initialize(ctx context.Context, db *gorm.DB, seeder PublicWorkbookSeeder, ownerLoginID string, ownerPassword string) error {
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
	if err := assignGroup(ctx, rbacRepo, org.ID(), ownerUserID, domainrbac.GroupAdmin(), logger); err != nil {
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
	if err := setupSystemOwnerAndSpace(ctx, db, seeder, rbacRepo, org.ID(), logger); err != nil {
		return fmt.Errorf("setup system owner and space: %w", err)
	}

	logger.InfoContext(ctx, "initialization completed successfully",
		slog.String("organization_id", org.ID().String()),
		slog.String("owner_user_id", ownerUserID.String()),
		slog.String("guest_user_id", guestUserID.String()),
	)
	return nil
}

// setupSystemOwnerAndSpace binds the system_owner group to the global
// SystemAppUserID instead of a per-tenant user: cocotola-question's API-key
// middleware stamps internal calls with that ID, so any other user fails authz.
func setupSystemOwnerAndSpace(ctx context.Context, db *gorm.DB, seeder PublicWorkbookSeeder, rbacRepo *gateway.RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	systemOwnerID := domain.SystemAppUserID()

	if err := setupSystemOwnerRBACPolicies(ctx, rbacRepo, orgID, logger); err != nil {
		return fmt.Errorf("setup system owner rbac policies: %w", err)
	}

	if err := assignGroup(ctx, rbacRepo, orgID, systemOwnerID, domainrbac.GroupSystemOwner(), logger); err != nil {
		return fmt.Errorf("assign system owner group: %w", err)
	}

	spaceRepo := gateway.NewSpaceRepository(db)
	publicSpace, err := findOrCreatePublicSpace(ctx, spaceRepo, orgID, systemOwnerID, logger)
	if err != nil {
		return fmt.Errorf("find or create public space: %w", err)
	}

	if err := grantPublicSpacePoliciesToSystemOwner(ctx, rbacRepo, orgID, publicSpace.ID(), logger); err != nil {
		return fmt.Errorf("grant public space policies to system owner: %w", err)
	}

	if err := seeder.SeedPublicWorkbooks(ctx, orgID.String(), publicSpace.ID().String()); err != nil {
		return fmt.Errorf("seed public workbooks: %w", err)
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

	org, err = domain.NewOrganization(CocotolaOrganizationID(), organizationName, maxActiveUsers, maxActiveGroups)
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

	user, err = domainuser.Provision(ctx, repo, orgID, domain.LoginID(loginID), hashedPassword, true)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("provision owner user: %w", err)
	}

	logger.InfoContext(ctx, "owner user created",
		slog.String("user_id", user.ID().String()),
	)
	return user.ID(), nil
}

func findOrCreateGuest(ctx context.Context, repo *gateway.AppUserRepository, orgID domain.OrganizationID, logger *slog.Logger) (domain.AppUserID, error) {
	guestLoginID := domainuser.NewGuestLoginID(organizationName)

	user, err := repo.FindByLoginID(ctx, orgID, domain.LoginID(guestLoginID))
	if err == nil {
		logger.InfoContext(ctx, "guest user already exists",
			slog.String("user_id", user.ID().String()),
		)
		return user.ID(), nil
	}
	if !errors.Is(err, domain.ErrAppUserNotFound) {
		return domain.AppUserID{}, fmt.Errorf("find guest by login id: %w", err)
	}

	user, err = domainuser.Provision(ctx, repo, orgID, domain.LoginID(guestLoginID), "", true)
	if err != nil {
		return domain.AppUserID{}, fmt.Errorf("provision guest user: %w", err)
	}

	logger.InfoContext(ctx, "guest user created",
		slog.String("user_id", user.ID().String()),
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

func adminActions() []domainrbac.Action {
	return []domainrbac.Action{
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
}

func setupRBACPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	return addGroupPolicies(ctx, repo, orgID, domainrbac.GroupAdmin(), domainrbac.ResourceAny(), adminActions(), logger)
}

func setupSystemOwnerRBACPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	actions := append(adminActions(), domainrbac.ActionCreateOrganization())

	return addGroupPolicies(ctx, repo, orgID, domainrbac.GroupSystemOwner(), domainrbac.ResourceAny(), actions, logger)
}

func addGroupPolicies(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, group domainrbac.Group, resource domainrbac.Resource, actions []domainrbac.Action, logger *slog.Logger) error {
	for _, action := range actions {
		if err := repo.AddPolicy(ctx, orgID, group, action, resource, domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add %s policy %s on %s: %w", group.Value(), action.Value(), resource.Value(), err)
		}
	}

	logger.InfoContext(ctx, "RBAC policies added",
		slog.String("organization_id", orgID.String()),
		slog.String("group", group.Value()),
		slog.String("resource", resource.Value()),
		slog.Int("policy_count", len(actions)),
	)
	return nil
}

func assignGroup(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, userID domain.AppUserID, group domainrbac.Group, logger *slog.Logger) error {
	if err := repo.AssignGroupToUser(ctx, orgID, userID, group); err != nil {
		return fmt.Errorf("user %s: %w", userID.String(), err)
	}

	logger.InfoContext(ctx, "RBAC group assigned to user",
		slog.String("organization_id", orgID.String()),
		slog.String("group", group.Value()),
		slog.String("user_id", userID.String()),
	)
	return nil
}

func findOrCreatePublicSpace(ctx context.Context, repo *gateway.SpaceRepository, orgID domain.OrganizationID, systemOwnerID domain.AppUserID, logger *slog.Logger) (*domainspace.Space, error) {
	keyName := domainspace.PublicSpaceKeyName(organizationName)

	existing, err := repo.FindByKeyName(ctx, orgID, keyName)
	if err == nil {
		logger.InfoContext(ctx, "public space already exists", slog.String("key_name", keyName))
		return existing, nil
	}
	if !errors.Is(err, domain.ErrSpaceNotFound) {
		return nil, fmt.Errorf("find public space by key name: %w", err)
	}

	s, err := domainspace.Provision(ctx, repo, orgID, systemOwnerID, keyName, "Public", domainspace.TypePublic())
	if err != nil {
		return nil, fmt.Errorf("provision public space: %w", err)
	}

	logger.InfoContext(ctx, "public space created",
		slog.String("space_id", s.ID().String()),
		slog.String("key_name", keyName),
	)
	return s, nil
}

func grantPublicSpacePoliciesToSystemOwner(ctx context.Context, repo *gateway.RBACRepository, orgID domain.OrganizationID, publicSpaceID domain.SpaceID, logger *slog.Logger) error {
	actions := []domainrbac.Action{
		domainrbac.ActionCreateWorkbook(),
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionUpdateWorkbook(),
		domainrbac.ActionDeleteWorkbook(),
		domainrbac.ActionCreateQuestion(),
		domainrbac.ActionUpdateQuestion(),
		domainrbac.ActionDeleteQuestion(),
	}

	return addGroupPolicies(ctx, repo, orgID, domainrbac.GroupSystemOwner(), domainrbac.ResourceSpace(publicSpaceID), actions, logger)
}
