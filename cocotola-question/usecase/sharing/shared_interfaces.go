package sharing

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type referenceSaver interface {
	Save(ctx context.Context, ref *domainreference.WorkbookReference) error
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
	FindPublicByOrganizationIDAndLanguage(ctx context.Context, organizationID string, language string) ([]domainworkbook.Workbook, error)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error)
}
