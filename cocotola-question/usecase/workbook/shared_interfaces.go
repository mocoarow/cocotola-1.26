package workbook

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type workbookCreator interface {
	Create(ctx context.Context, spaceID int, ownerID string, organizationID string, title string, description string, visibility string) (string, error)
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
	FindBySpaceID(ctx context.Context, spaceID int) ([]domainworkbook.Workbook, error)
}

type workbookUpdater interface {
	Update(ctx context.Context, wb *domainworkbook.Workbook) error
}

type workbookDeleter interface {
	Delete(ctx context.Context, id string) error
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error)
}
