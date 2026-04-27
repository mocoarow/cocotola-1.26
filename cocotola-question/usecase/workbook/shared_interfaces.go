package workbook

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type workbookCreator interface {
	Create(ctx context.Context, spaceID string, ownerID string, organizationID string, title string, description string, visibility string, language string) (string, error)
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
	FindBySpaceID(ctx context.Context, spaceID string) ([]domainworkbook.Workbook, error)
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

type policyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID string, userID string, action domain.Action, resource domain.Resource, effect domain.Effect) error
}

type ownedWorkbookListFinder interface {
	FindByOwnerID(ctx context.Context, ownerID string) (*domain.OwnedWorkbookList, error)
}

type ownedWorkbookListSaver interface {
	Save(ctx context.Context, list *domain.OwnedWorkbookList) error
}

type maxWorkbooksFetcher interface {
	FetchMaxWorkbooks(ctx context.Context, userID string) (int, error)
}

// spaceTypeFetcher resolves a space's type ("public" or "private") via cocotola-auth.
type spaceTypeFetcher interface {
	FetchSpaceType(ctx context.Context, spaceID string) (string, error)
}
