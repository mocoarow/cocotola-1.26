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

// RBACRepository manages RBAC policies and role assignments using Casbin.
// It implements domain.RBACPolicyRepository.
type RBACRepository struct {
	enforcer *casbin.Enforcer
}

var _ domain.RBACPolicyRepository = (*RBACRepository)(nil)

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

func formatDomain(organizationID int) string {
	return fmt.Sprintf("org:%d", organizationID)
}

func formatSubject(userID int) string {
	return fmt.Sprintf("user:%d", userID)
}

// AssignRoleToUser assigns a role to a user within an organization.
func (r *RBACRepository) AssignRoleToUser(_ context.Context, organizationID int, userID int, role domain.RBACRole) error {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	if _, err := r.enforcer.AddNamedGroupingPolicy("g", sub, role.Value(), dom); err != nil {
		return fmt.Errorf("add grouping policy: %w", err)
	}

	return nil
}

// RevokeRoleFromUser revokes a role from a user within an organization.
func (r *RBACRepository) RevokeRoleFromUser(_ context.Context, organizationID int, userID int, role domain.RBACRole) error {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	if _, err := r.enforcer.RemoveNamedGroupingPolicy("g", sub, role.Value(), dom); err != nil {
		return fmt.Errorf("remove grouping policy: %w", err)
	}

	return nil
}

// AddPolicy adds a policy rule granting or denying a role an action on a resource.
func (r *RBACRepository) AddPolicy(_ context.Context, organizationID int, role domain.RBACRole, action domain.RBACAction, resource domain.RBACResource, effect domain.RBACEffect) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.AddNamedPolicy("p", role.Value(), resource.Value(), action.Value(), effect.Value(), dom); err != nil {
		return fmt.Errorf("add named policy: %w", err)
	}

	return nil
}

// RemovePolicy removes a policy rule.
func (r *RBACRepository) RemovePolicy(_ context.Context, organizationID int, role domain.RBACRole, action domain.RBACAction, resource domain.RBACResource, effect domain.RBACEffect) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.RemoveNamedPolicy("p", role.Value(), resource.Value(), action.Value(), effect.Value(), dom); err != nil {
		return fmt.Errorf("remove named policy: %w", err)
	}

	return nil
}

// AddObjectGroupingPolicy adds a parent-child relationship between resources.
func (r *RBACRepository) AddObjectGroupingPolicy(_ context.Context, organizationID int, child domain.RBACResource, parent domain.RBACResource) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.AddNamedGroupingPolicy("g2", child.Value(), parent.Value(), dom); err != nil {
		return fmt.Errorf("add object grouping policy: %w", err)
	}

	return nil
}

// RemoveObjectGroupingPolicy removes a parent-child relationship between resources.
func (r *RBACRepository) RemoveObjectGroupingPolicy(_ context.Context, organizationID int, child domain.RBACResource, parent domain.RBACResource) error {
	dom := formatDomain(organizationID)

	if _, err := r.enforcer.RemoveNamedGroupingPolicy("g2", child.Value(), parent.Value(), dom); err != nil {
		return fmt.Errorf("remove object grouping policy: %w", err)
	}

	return nil
}

// Enforce checks whether a subject is allowed to perform an action on a resource.
func (r *RBACRepository) Enforce(organizationID int, userID int, action domain.RBACAction, resource domain.RBACResource) (bool, error) {
	dom := formatDomain(organizationID)
	sub := formatSubject(userID)

	ok, err := r.enforcer.Enforce(sub, resource.Value(), action.Value(), dom)
	if err != nil {
		return false, fmt.Errorf("enforce: %w", err)
	}

	return ok, nil
}

// LoadPolicy reloads all policies from the storage.
func (r *RBACRepository) LoadPolicy() error {
	if err := r.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("load policy: %w", err)
	}
	return nil
}
