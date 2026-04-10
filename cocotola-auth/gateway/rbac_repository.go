package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
)

const rbacModelConf = `
[request_definition]
r = sub, obj, act, dom

[policy_definition]
p = sub, obj, act, eft, dom

[role_definition]
g = _, _, _
g2 = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && (keyMatch(r.obj, p.obj) || g2(r.obj, p.obj, r.dom)) && r.act == p.act
`

// RBACRepository manages RBAC policies and group assignments using Casbin.
// It implements domainrbac.PolicyRepository and domainrbac.GroupFinder.
type RBACRepository struct {
	enforcer *casbin.Enforcer
}

var _ domainrbac.PolicyRepository = (*RBACRepository)(nil)
var _ domainrbac.GroupFinder = (*RBACRepository)(nil)
var _ domainrbac.UserPolicyManager = (*RBACRepository)(nil)

// NewRBACRepository creates a new RBACRepository backed by the given database.
func NewRBACRepository(db *gorm.DB) (*RBACRepository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	gormadapter.TurnOffAutoMigrate(db)

	a, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("new gorm adapter: %w", err)
	}

	m, err := model.NewModelFromString(rbacModelConf)
	if err != nil {
		return nil, fmt.Errorf("new model from string: %w", err)
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, fmt.Errorf("new enforcer: %w", err)
	}
	e.EnableAutoSave(true)

	return &RBACRepository{
		enforcer: e,
	}, nil
}

func formatDomain(organizationID domain.OrganizationID) string {
	return "org:" + organizationID.String()
}

func formatSubject(userID domain.AppUserID) string {
	return "user:" + userID.String()
}

// AssignGroupToUser assigns a group to a user within an organization.
func (r *RBACRepository) AssignGroupToUser(_ context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, group domainrbac.Group) error {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	if _, err := r.enforcer.AddNamedGroupingPolicy("g", sub, group.Value(), dom); err != nil {
		return fmt.Errorf("add grouping policy: %w", err)
	}

	return nil
}

// RevokeGroupFromUser revokes a group from a user within an organization.
func (r *RBACRepository) RevokeGroupFromUser(_ context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, group domainrbac.Group) error {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	if _, err := r.enforcer.RemoveNamedGroupingPolicy("g", sub, group.Value(), dom); err != nil {
		return fmt.Errorf("remove grouping policy: %w", err)
	}

	return nil
}

// AddPolicy adds a policy rule granting or denying a group an action on a resource.
func (r *RBACRepository) AddPolicy(_ context.Context, organizationID domain.OrganizationID, group domainrbac.Group, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.AddNamedPolicy("p", group.Value(), resource.Value(), action.Value(), effect.Value(), dom); err != nil {
		return fmt.Errorf("add named policy: %w", err)
	}

	return nil
}

// RemovePolicy removes a policy rule.
func (r *RBACRepository) RemovePolicy(_ context.Context, organizationID domain.OrganizationID, group domainrbac.Group, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.RemoveNamedPolicy("p", group.Value(), resource.Value(), action.Value(), effect.Value(), dom); err != nil {
		return fmt.Errorf("remove named policy: %w", err)
	}

	return nil
}

// AddPolicyForUser adds a policy rule for a specific user (not group).
func (r *RBACRepository) AddPolicyForUser(_ context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	if _, err := r.enforcer.AddNamedPolicy("p", sub, resource.Value(), action.Value(), effect.Value(), dom); err != nil {
		return fmt.Errorf("add user policy: %w", err)
	}

	return nil
}

// AddObjectGroupingPolicy adds a parent-child relationship between resources.
func (r *RBACRepository) AddObjectGroupingPolicy(_ context.Context, organizationID domain.OrganizationID, child domainrbac.Resource, parent domainrbac.Resource) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.AddNamedGroupingPolicy("g2", child.Value(), parent.Value(), dom); err != nil {
		return fmt.Errorf("add object grouping policy: %w", err)
	}

	return nil
}

// RemoveObjectGroupingPolicy removes a parent-child relationship between resources.
func (r *RBACRepository) RemoveObjectGroupingPolicy(_ context.Context, organizationID domain.OrganizationID, child domainrbac.Resource, parent domainrbac.Resource) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.RemoveNamedGroupingPolicy("g2", child.Value(), parent.Value(), dom); err != nil {
		return fmt.Errorf("remove object grouping policy: %w", err)
	}

	return nil
}

// Enforce checks whether a subject is allowed to perform an action on a resource.
func (r *RBACRepository) Enforce(organizationID domain.OrganizationID, userID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource) (bool, error) {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	ok, err := r.enforcer.Enforce(sub, resource.Value(), action.Value(), dom)
	if err != nil {
		return false, fmt.Errorf("enforce: %w", err)
	}

	return ok, nil
}

// GetGroupsForUser retrieves groups assigned to a user within an organization.
func (r *RBACRepository) GetGroupsForUser(_ context.Context, organizationID domain.OrganizationID, userID domain.AppUserID) ([]string, error) {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	groups, err := r.enforcer.GetRolesForUser(sub, dom)
	if err != nil {
		return nil, fmt.Errorf("get groups for user: %w", err)
	}

	return groups, nil
}

// LoadPolicy reloads all policies from the storage.
func (r *RBACRepository) LoadPolicy() error {
	if err := r.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("load policy: %w", err)
	}
	return nil
}
