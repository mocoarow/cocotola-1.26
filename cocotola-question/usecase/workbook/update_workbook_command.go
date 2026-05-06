package workbook

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// UpdateWorkbookCommand handles workbook updates.
type UpdateWorkbookCommand struct {
	workbookFinder workbookFinder
	workbookSaver  workbookSaver
	authChecker    authorizationChecker
}

// NewUpdateWorkbookCommand returns a new UpdateWorkbookCommand.
func NewUpdateWorkbookCommand(workbookFinder workbookFinder, workbookSaver workbookSaver, authChecker authorizationChecker) *UpdateWorkbookCommand {
	return &UpdateWorkbookCommand{
		workbookFinder: workbookFinder,
		workbookSaver:  workbookSaver,
		authChecker:    authChecker,
	}
}

// UpdateWorkbook updates an existing workbook.
func (c *UpdateWorkbookCommand) UpdateWorkbook(ctx context.Context, input *workbookservice.UpdateWorkbookInput) (*workbookservice.UpdateWorkbookOutput, error) {
	resource, err := domain.ResourceWorkbook(input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("resource workbook: %w", err)
	}
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionUpdateWorkbook(), resource)
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	wb, err := c.workbookFinder.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if wb.OwnerID() != input.OperatorID {
		return nil, domain.ErrForbidden
	}

	if err := wb.UpdateTitle(input.Title); err != nil {
		return nil, fmt.Errorf("update title: %w", err)
	}
	if err := wb.UpdateDescription(input.Description); err != nil {
		return nil, fmt.Errorf("update description: %w", err)
	}

	vis, err := domainworkbook.NewVisibility(input.Visibility)
	if err != nil {
		return nil, fmt.Errorf("new visibility: %w", err)
	}
	wb.ChangeVisibility(vis)

	lang, err := domainworkbook.NewLanguage(input.Language)
	if err != nil {
		return nil, fmt.Errorf("new language: %w", err)
	}
	wb.ChangeLanguage(lang)

	wb.Touch(time.Now())

	if err := c.workbookSaver.Save(ctx, wb); err != nil {
		return nil, fmt.Errorf("save workbook: %w", err)
	}

	return &workbookservice.UpdateWorkbookOutput{
		Item: workbookservice.Item{
			WorkbookID:     wb.ID(),
			SpaceID:        wb.SpaceID(),
			OwnerID:        wb.OwnerID(),
			OrganizationID: wb.OrganizationID(),
			Title:          wb.Title(),
			Description:    wb.Description(),
			Visibility:     wb.Visibility().Value(),
			Language:       wb.Language().Value(),
			CreatedAt:      wb.CreatedAt(),
			UpdatedAt:      wb.UpdatedAt(),
		},
	}, nil
}
