package workbook

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// CreateWorkbookCommand handles workbook creation.
type CreateWorkbookCommand struct {
	workbookRepo workbookCreator
	authChecker  authorizationChecker
}

// NewCreateWorkbookCommand returns a new CreateWorkbookCommand.
func NewCreateWorkbookCommand(workbookRepo workbookCreator, authChecker authorizationChecker) *CreateWorkbookCommand {
	return &CreateWorkbookCommand{
		workbookRepo: workbookRepo,
		authChecker:  authChecker,
	}
}

// CreateWorkbook creates a new workbook.
func (c *CreateWorkbookCommand) CreateWorkbook(ctx context.Context, input *workbookservice.CreateWorkbookInput) (*workbookservice.CreateWorkbookOutput, error) {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	if _, err := domainworkbook.NewVisibility(input.Visibility); err != nil {
		return nil, fmt.Errorf("new visibility: %w", err)
	}

	workbookID, err := c.workbookRepo.Create(ctx, input.SpaceID, input.OperatorID, input.OrganizationID, input.Title, input.Description, input.Visibility)
	if err != nil {
		return nil, fmt.Errorf("create workbook: %w", err)
	}

	now := time.Now()
	output, err := workbookservice.NewCreateWorkbookOutput(workbookID, input.SpaceID, input.OperatorID, input.OrganizationID, input.Title, input.Description, input.Visibility, now, now)
	if err != nil {
		return nil, fmt.Errorf("create workbook output: %w", err)
	}
	return output, nil
}
