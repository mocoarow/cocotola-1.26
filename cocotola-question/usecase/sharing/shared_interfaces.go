package sharing

import (
	"context"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type referenceCreator interface {
	Create(ctx context.Context, userID int, workbookID string) (string, error)
}

type referenceFinder interface {
	FindByID(ctx context.Context, userID int, referenceID string) (*domainreference.WorkbookReference, error)
	FindByUserID(ctx context.Context, userID int) ([]domainreference.WorkbookReference, error)
}

type referenceDeleter interface {
	Delete(ctx context.Context, userID int, referenceID string) error
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
}

type publicWorkbookFinder interface {
	FindPublicByOrganizationID(ctx context.Context, organizationID int) ([]domainworkbook.Workbook, error)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}
