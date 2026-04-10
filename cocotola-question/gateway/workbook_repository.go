package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

const workbooksCollection = "workbooks"

type workbookRecord struct {
	SpaceID        string    `firestore:"spaceID"`
	OwnerID        string    `firestore:"ownerID"`
	OrganizationID string    `firestore:"organizationID"`
	Title          string    `firestore:"title"`
	Description    string    `firestore:"description"`
	Visibility     string    `firestore:"visibility"`
	CreatedAt      time.Time `firestore:"createdAt"`
	UpdatedAt      time.Time `firestore:"updatedAt"`
}

func toWorkbookDomain(id string, r *workbookRecord) (*domainworkbook.Workbook, error) {
	vis, err := domainworkbook.NewVisibility(r.Visibility)
	if err != nil {
		return nil, fmt.Errorf("invalid visibility %q: %w", r.Visibility, err)
	}
	return domainworkbook.ReconstructWorkbook(id, r.SpaceID, r.OwnerID, r.OrganizationID, r.Title, r.Description, vis, r.CreatedAt, r.UpdatedAt), nil
}

// WorkbookRepository manages workbook persistence in Firestore.
type WorkbookRepository struct {
	client *firestore.Client
}

// NewWorkbookRepository returns a new WorkbookRepository.
func NewWorkbookRepository(client *firestore.Client) *WorkbookRepository {
	return &WorkbookRepository{client: client}
}

// Create inserts a new workbook and returns the auto-generated document ID.
func (r *WorkbookRepository) Create(ctx context.Context, spaceID string, ownerID string, organizationID string, title string, description string, visibility string) (string, error) {
	now := time.Now()
	record := workbookRecord{
		SpaceID:        spaceID,
		OwnerID:        ownerID,
		OrganizationID: organizationID,
		Title:          title,
		Description:    description,
		Visibility:     visibility,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	docRef, _, err := r.client.Collection(workbooksCollection).Add(ctx, record)
	if err != nil {
		return "", fmt.Errorf("create workbook: %w", err)
	}
	return docRef.ID, nil
}

// FindByID looks up a workbook by its document ID.
func (r *WorkbookRepository) FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error) {
	doc, err := r.client.Collection(workbooksCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrWorkbookNotFound
		}
		return nil, fmt.Errorf("find workbook by id: %w", err)
	}
	var record workbookRecord
	if err := doc.DataTo(&record); err != nil {
		return nil, fmt.Errorf("decode workbook: %w", err)
	}
	wb, err := toWorkbookDomain(doc.Ref.ID, &record)
	if err != nil {
		return nil, fmt.Errorf("convert workbook domain: %w", err)
	}
	return wb, nil
}

// FindBySpaceID returns all workbooks for the given space.
func (r *WorkbookRepository) FindBySpaceID(ctx context.Context, spaceID string) ([]domainworkbook.Workbook, error) {
	iter := r.client.Collection(workbooksCollection).Where("spaceID", "==", spaceID).Documents(ctx)
	defer iter.Stop()

	var workbooks []domainworkbook.Workbook

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate workbooks: %w", err)
		}
		var record workbookRecord
		if err := doc.DataTo(&record); err != nil {
			return nil, fmt.Errorf("decode workbook: %w", err)
		}
		wb, err := toWorkbookDomain(doc.Ref.ID, &record)
		if err != nil {
			return nil, fmt.Errorf("convert workbook domain: %w", err)
		}
		workbooks = append(workbooks, *wb)
	}
	return workbooks, nil
}

// FindPublicByOrganizationID returns all public workbooks for the given organization.
func (r *WorkbookRepository) FindPublicByOrganizationID(ctx context.Context, organizationID string) ([]domainworkbook.Workbook, error) {
	iter := r.client.Collection(workbooksCollection).
		Where("organizationID", "==", organizationID).
		Where("visibility", "==", "public").
		Documents(ctx)
	defer iter.Stop()

	var workbooks []domainworkbook.Workbook

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate public workbooks: %w", err)
		}
		var record workbookRecord
		if err := doc.DataTo(&record); err != nil {
			return nil, fmt.Errorf("decode workbook: %w", err)
		}
		wb, err := toWorkbookDomain(doc.Ref.ID, &record)
		if err != nil {
			return nil, fmt.Errorf("convert workbook domain: %w", err)
		}
		workbooks = append(workbooks, *wb)
	}
	return workbooks, nil
}

// Update updates an existing workbook.
func (r *WorkbookRepository) Update(ctx context.Context, wb *domainworkbook.Workbook) error {
	now := time.Now()
	_, err := r.client.Collection(workbooksCollection).Doc(wb.ID()).Set(ctx, workbookRecord{
		SpaceID:        wb.SpaceID(),
		OwnerID:        wb.OwnerID(),
		OrganizationID: wb.OrganizationID(),
		Title:          wb.Title(),
		Description:    wb.Description(),
		Visibility:     wb.Visibility().Value(),
		CreatedAt:      wb.CreatedAt(),
		UpdatedAt:      now,
	})
	if err != nil {
		return fmt.Errorf("update workbook: %w", err)
	}
	return nil
}

// Delete removes a workbook document.
func (r *WorkbookRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(workbooksCollection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete workbook: %w", err)
	}
	return nil
}
