package initialize

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

// studyWorkbookAction must match the literal cocotola-question grants;
// cocotola-auth's domain has no constructor for it.
const studyWorkbookAction = "study_workbook"

// WorkbookPolicyEnsurer writes SystemAppUserID's per-workbook Casbin policies
// directly to the auth database, mirroring cocotola-question's grants.
type WorkbookPolicyEnsurer struct {
	rbacRepo *gateway.RBACRepository
}

// NewWorkbookPolicyEnsurer constructs a WorkbookPolicyEnsurer backed by the
// auth database.
func NewWorkbookPolicyEnsurer(db *gorm.DB) (*WorkbookPolicyEnsurer, error) {
	rbacRepo, err := gateway.NewRBACRepository(db)
	if err != nil {
		return nil, fmt.Errorf("new rbac repository: %w", err)
	}
	return &WorkbookPolicyEnsurer{rbacRepo: rbacRepo}, nil
}

// EnsureSystemOwnerWorkbookPolicies idempotently grants SystemAppUserID the
// workbook-scoped policies on the given workbook, so repeated init runs are safe
// and orphaned workbooks are repaired in place.
func (e *WorkbookPolicyEnsurer) EnsureSystemOwnerWorkbookPolicies(ctx context.Context, organizationID, workbookID string) error {
	orgID, err := domain.ParseOrganizationID(organizationID)
	if err != nil {
		return fmt.Errorf("parse organization id %q: %w", organizationID, err)
	}

	resource, err := domainrbac.ResourceWorkbook(workbookID)
	if err != nil {
		return fmt.Errorf("resource workbook %q: %w", workbookID, err)
	}

	studyWorkbook, err := domainrbac.NewAction(studyWorkbookAction)
	if err != nil {
		return fmt.Errorf("new study_workbook action: %w", err)
	}

	actions := []domainrbac.Action{
		domainrbac.ActionViewWorkbook(),
		domainrbac.ActionUpdateWorkbook(),
		domainrbac.ActionDeleteWorkbook(),
		studyWorkbook,
		domainrbac.ActionCreateQuestion(),
		domainrbac.ActionUpdateQuestion(),
		domainrbac.ActionDeleteQuestion(),
	}

	systemOwnerID := domain.SystemAppUserID()
	for _, action := range actions {
		if err := e.rbacRepo.AddPolicyForUser(ctx, orgID, systemOwnerID, action, resource, domainrbac.EffectAllow()); err != nil {
			return fmt.Errorf("add system owner workbook policy %s on workbook %s: %w", action.Value(), workbookID, err)
		}
	}

	return nil
}
