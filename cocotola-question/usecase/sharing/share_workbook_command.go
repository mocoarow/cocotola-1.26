package sharing

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

// ShareWorkbookCommand handles importing a workbook reference.
type ShareWorkbookCommand struct {
	referenceRepo referenceCreator
	workbookRepo  workbookFinder
	authChecker   authorizationChecker
}

// NewShareWorkbookCommand returns a new ShareWorkbookCommand.
func NewShareWorkbookCommand(referenceRepo referenceCreator, workbookRepo workbookFinder, authChecker authorizationChecker) *ShareWorkbookCommand {
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

	refID, err := c.referenceRepo.Create(ctx, input.OperatorID, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("create reference: %w", err)
	}

	return &referenceservice.ShareWorkbookOutput{
		ReferenceID: refID,
		WorkbookID:  input.WorkbookID,
		AddedAt:     time.Now(),
	}, nil
}
