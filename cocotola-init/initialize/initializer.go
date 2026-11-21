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
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"
)

// PublicWorkbookSeeder seeds public workbooks (and their questions) into
// cocotola-question via its internal HTTP API.
type PublicWorkbookSeeder interface {
	SeedPublicWorkbooks(ctx context.Context, organizationID, publicSpaceID string) error
}

// OrganizationRepository finds and saves organizations.
type OrganizationRepository interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
	Save(ctx context.Context, org *domain.Organization) error
}

// AppUserRepository finds and saves app users.
type AppUserRepository interface {
	FindByLoginID(ctx context.Context, organizationID domain.OrganizationID, loginID domain.LoginID) (*domainuser.AppUser, error)
	Save(ctx context.Context, user *domainuser.AppUser) error
}

// ActiveUserListRepository finds and saves an organization's active user list.
type ActiveUserListRepository interface {
	FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*domain.ActiveUserList, error)
	Save(ctx context.Context, list *domain.ActiveUserList) error
}

// RBACRepository adds group policies and assigns groups to users.
type RBACRepository interface {
	AddPolicy(ctx context.Context, organizationID domain.OrganizationID, group domainrbac.Group, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error
	AssignGroupToUser(ctx context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, group domainrbac.Group) error
}

// SpaceRepository finds and saves spaces.
type SpaceRepository interface {
	FindByKeyName(ctx context.Context, organizationID domain.OrganizationID, keyName string) (*domainspace.Space, error)
	Save(ctx context.Context, space *domainspace.Space) error
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

	rbacRepo, err := gateway.NewRBACRepository(db)
	if err != nil {
		return fmt.Errorf("new rbac repository: %w", err)
	}

	users, err := bootstrapTenantUsers(ctx, db, rbacRepo, ownerLoginID, ownerPassword, logger)
	if err != nil {
		return fmt.Errorf("bootstrap tenant users: %w", err)
	}

	if err := setupSystemOwnerAndSpace(ctx, db, seeder, rbacRepo, users.orgID, logger); err != nil {
		return fmt.Errorf("setup system owner and space: %w", err)
	}

	logger.InfoContext(ctx, "initialization completed successfully",
		slog.String("organization_id", users.orgID.String()),
		slog.String("owner_user_id", users.ownerUserID.String()),
		slog.String("guest_user_id", users.guestUserID.String()),
	)
	return nil
}

type tenantUsers struct {
	orgID       domain.OrganizationID
	ownerUserID domain.AppUserID
	guestUserID domain.AppUserID
}

func bootstrapTenantUsers(ctx context.Context, db *gorm.DB, rbacRepo *gateway.RBACRepository, ownerLoginID string, ownerPassword string, logger *slog.Logger) (tenantUsers, error) {
	org, err := findOrCreateOrganization(ctx, gateway.NewOrganizationRepository(db), logger)
	if err != nil {
		return tenantUsers{}, fmt.Errorf("find or create organization: %w", err)
	}

	appUserRepo := gateway.NewAppUserRepository(db)
	ownerUserID, err := findOrCreateOwner(ctx, appUserRepo, gateway.NewBcryptHasher(), org.ID(), ownerLoginID, ownerPassword, logger)
	if err != nil {
		return tenantUsers{}, fmt.Errorf("find or create owner: %w", err)
	}

	activeUserListRepo := gateway.NewActiveUserListRepository(db)
	if err := addToActiveUserList(ctx, activeUserListRepo, org, ownerUserID, logger); err != nil {
		return tenantUsers{}, fmt.Errorf("owner: %w", err)
	}

	if err := setupRBACPolicies(ctx, rbacRepo, org.ID(), logger); err != nil {
		return tenantUsers{}, fmt.Errorf("setup rbac policies: %w", err)
	}

	if err := assignGroup(ctx, rbacRepo, org.ID(), ownerUserID, domainrbac.GroupAdmin(), logger); err != nil {
		return tenantUsers{}, fmt.Errorf("assign admin group: %w", err)
	}

	guestUserID, err := findOrCreateGuest(ctx, appUserRepo, org.ID(), logger)
	if err != nil {
		return tenantUsers{}, fmt.Errorf("find or create guest: %w", err)
	}

	if err := addToActiveUserList(ctx, activeUserListRepo, org, guestUserID, logger); err != nil {
		return tenantUsers{}, fmt.Errorf("guest: %w", err)
	}

	return tenantUsers{orgID: org.ID(), ownerUserID: ownerUserID, guestUserID: guestUserID}, nil
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

func findOrCreateOrganization(ctx context.Context, repo OrganizationRepository, logger *slog.Logger) (*domain.Organization, error) {
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

func findOrCreateOwner(ctx context.Context, repo AppUserRepository, hasher domainuser.PasswordHasher, orgID domain.OrganizationID, loginID string, rawPassword string, logger *slog.Logger) (domain.AppUserID, error) {
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

func findOrCreateGuest(ctx context.Context, repo AppUserRepository, orgID domain.OrganizationID, logger *slog.Logger) (domain.AppUserID, error) {
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

func addToActiveUserList(ctx context.Context, repo ActiveUserListRepository, org *domain.Organization, userID domain.AppUserID, logger *slog.Logger) error {
	err := occ.Retry(func() error {
		list, err := repo.FindByOrganizationID(ctx, org.ID())
		if err != nil {
			return fmt.Errorf("find: %w", err)
		}
		if err := list.Add(userID, org.MaxActiveUsers()); err != nil {
			return fmt.Errorf("add: %w", err)
		}
		if err := repo.Save(ctx, list); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		return nil
	})
	if errors.Is(err, domain.ErrDuplicateEntry) {
		logger.InfoContext(ctx, "user already in active user list", slog.String("user_id", userID.String()))
		return nil
	}
	if err != nil {
		return fmt.Errorf("add user %s to active user list: %w", userID, err)
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

func setupRBACPolicies(ctx context.Context, repo RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	return addGroupPolicies(ctx, repo, orgID, domainrbac.GroupAdmin(), domainrbac.ResourceAny(), adminActions(), logger)
}

func setupSystemOwnerRBACPolicies(ctx context.Context, repo RBACRepository, orgID domain.OrganizationID, logger *slog.Logger) error {
	actions := append(adminActions(), domainrbac.ActionCreateOrganization())

	return addGroupPolicies(ctx, repo, orgID, domainrbac.GroupSystemOwner(), domainrbac.ResourceAny(), actions, logger)
}

func addGroupPolicies(ctx context.Context, repo RBACRepository, orgID domain.OrganizationID, group domainrbac.Group, resource domainrbac.Resource, actions []domainrbac.Action, logger *slog.Logger) error {
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

func assignGroup(ctx context.Context, repo RBACRepository, orgID domain.OrganizationID, userID domain.AppUserID, group domainrbac.Group, logger *slog.Logger) error {
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

func findOrCreatePublicSpace(ctx context.Context, repo SpaceRepository, orgID domain.OrganizationID, systemOwnerID domain.AppUserID, logger *slog.Logger) (*domainspace.Space, error) {
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

func grantPublicSpacePoliciesToSystemOwner(ctx context.Context, repo RBACRepository, orgID domain.OrganizationID, publicSpaceID domain.SpaceID, logger *slog.Logger) error {
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
