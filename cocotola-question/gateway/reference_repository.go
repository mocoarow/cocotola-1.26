package gateway

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
)

const (
	usersCollection    = "users"
	workbookRefsSubCol = "workbook_refs"
)

type referenceRecord struct {
	WorkbookID string    `firestore:"workbookID"`
	AddedAt    time.Time `firestore:"addedAt"`
}

func toReferenceDomain(id string, userID int, r *referenceRecord) *domainreference.WorkbookReference {
	return domainreference.ReconstructWorkbookReference(id, userID, r.WorkbookID, r.AddedAt)
}

// ReferenceRepository manages workbook reference persistence in Firestore.
type ReferenceRepository struct {
	client *firestore.Client
}

// NewReferenceRepository returns a new ReferenceRepository.
func NewReferenceRepository(client *firestore.Client) *ReferenceRepository {
	return &ReferenceRepository{client: client}
}

func (r *ReferenceRepository) refsCol(userID int) *firestore.CollectionRef {
	return r.client.Collection(usersCollection).Doc(strconv.Itoa(userID)).Collection(workbookRefsSubCol)
}

// Create inserts a new workbook reference and returns the auto-generated document ID.
func (r *ReferenceRepository) Create(ctx context.Context, userID int, workbookID string) (string, error) {
	// Check for duplicate
	iter := r.refsCol(userID).Where("workbookID", "==", workbookID).Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == nil && doc != nil {
		return "", domain.ErrDuplicateReference
	}

	record := referenceRecord{
		WorkbookID: workbookID,
		AddedAt:    time.Now(),
	}
	docRef, _, err := r.refsCol(userID).Add(ctx, record)
	if err != nil {
		return "", fmt.Errorf("create reference: %w", err)
	}
	return docRef.ID, nil
}

// FindByID looks up a reference by user ID and reference ID.
func (r *ReferenceRepository) FindByID(ctx context.Context, userID int, referenceID string) (*domainreference.WorkbookReference, error) {
	doc, err := r.refsCol(userID).Doc(referenceID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrReferenceNotFound
		}
		return nil, fmt.Errorf("find reference by id: %w", err)
	}
	var record referenceRecord
	if err := doc.DataTo(&record); err != nil {
		return nil, fmt.Errorf("decode reference: %w", err)
	}
	return toReferenceDomain(doc.Ref.ID, userID, &record), nil
}

// FindByUserID returns all references for the given user.
func (r *ReferenceRepository) FindByUserID(ctx context.Context, userID int) ([]domainreference.WorkbookReference, error) {
	iter := r.refsCol(userID).Documents(ctx)
	defer iter.Stop()

	var refs []domainreference.WorkbookReference

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate references: %w", err)
		}
		var record referenceRecord
		if err := doc.DataTo(&record); err != nil {
			return nil, fmt.Errorf("decode reference: %w", err)
		}
		refs = append(refs, *toReferenceDomain(doc.Ref.ID, userID, &record))
	}
	return refs, nil
}

// Delete removes a reference document.
func (r *ReferenceRepository) Delete(ctx context.Context, userID int, referenceID string) error {
	_, err := r.refsCol(userID).Doc(referenceID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete reference: %w", err)
	}
	return nil
}
