package sharing

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

// ShareWorkbookCommand handles importing a workbook reference.
type ShareWorkbookCommand struct {
	referenceRepo referenceSaver
	workbookRepo  workbookFinder
	authChecker   authorizationChecker
}

// NewShareWorkbookCommand returns a new ShareWorkbookCommand.
func NewShareWorkbookCommand(referenceRepo referenceSaver, workbookRepo workbookFinder, authChecker authorizationChecker) *ShareWorkbookCommand {
	return &ShareWorkbookCommand{
		referenceRepo: referenceRepo,
		workbookRepo:  workbookRepo,
		authChecker:   authChecker,
	}
}

// ShareWorkbook creates a reference to a workbook.
func (c *ShareWorkbookCommand) ShareWorkbook(ctx context.Context, input *referenceservice.ShareWorkbookInput) (*referenceservice.ShareWorkbookOutput, error) {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionImportWorkbook(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	wb, err := c.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if !wb.Visibility().IsPublic() {
		return nil, domain.ErrForbidden
	}

	addedAt := time.Now()
	ref, err := domainreference.NewWorkbookReference(input.OperatorID, input.WorkbookID, addedAt)
	if err != nil {
		return nil, fmt.Errorf("new workbook reference: %w", err)
	}

	if err := c.referenceRepo.Save(ctx, ref); err != nil {
		return nil, fmt.Errorf("save reference: %w", err)
	}

	return &referenceservice.ShareWorkbookOutput{
		ReferenceID: ref.ID(),
		WorkbookID:  ref.WorkbookID(),
		AddedAt:     ref.AddedAt(),
	}, nil
}
