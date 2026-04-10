package sharing

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type referenceCreator interface {
	Create(ctx context.Context, userID string, workbookID string) (string, error)
}

type referenceFinder interface {
	FindByID(ctx context.Context, userID string, referenceID string) (*domainreference.WorkbookReference, error)
	FindByUserID(ctx context.Context, userID string) ([]domainreference.WorkbookReference, error)
}

type referenceDeleter interface {
	Delete(ctx context.Context, userID string, referenceID string) error
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
}

type publicWorkbookFinder interface {
	FindPublicByOrganizationID(ctx context.Context, organizationID string) ([]domainworkbook.Workbook, error)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error)
}
