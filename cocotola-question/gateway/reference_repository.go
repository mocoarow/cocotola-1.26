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

func toReferenceDomain(userID string, r *referenceRecord) *domainreference.WorkbookReference {
	return domainreference.ReconstructWorkbookReference(userID, r.WorkbookID, r.AddedAt)
}

// ReferenceRepository manages workbook reference persistence in Firestore.
type ReferenceRepository struct {
	client *firestore.Client
}

// NewReferenceRepository returns a new ReferenceRepository.
func NewReferenceRepository(client *firestore.Client) *ReferenceRepository {
	return &ReferenceRepository{client: client}
}

func (r *ReferenceRepository) refsCol(userID string) *firestore.CollectionRef {
	return r.client.Collection(usersCollection).Doc(userID).Collection(workbookRefsSubCol)
}

// Save persists a workbook reference. The document is keyed by the referenced
// workbookID under the user's subcollection so the (userID, workbookID)
// uniqueness invariant is enforced atomically by Firestore: a second Save for
// the same pair fails with codes.AlreadyExists, which is mapped to
// domain.ErrDuplicateReference. This avoids the read-then-write race that an
// out-of-transaction duplicate-check query would have.
func (r *ReferenceRepository) Save(ctx context.Context, ref *domainreference.WorkbookReference) error {
	docRef := r.refsCol(ref.UserID()).Doc(ref.WorkbookID())
	record := referenceRecord{
		WorkbookID: ref.WorkbookID(),
		AddedAt:    ref.AddedAt(),
	}
	if _, err := docRef.Create(ctx, record); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return domain.ErrDuplicateReference
		}
		return fmt.Errorf("save reference: %w", err)
	}
	return nil
}

// FindByID looks up a reference by user ID and reference ID. Because a
// reference's ID is its workbookID, this is a direct doc lookup keyed by
// workbookID under the user's subcollection.
func (r *ReferenceRepository) FindByID(ctx context.Context, userID string, referenceID string) (*domainreference.WorkbookReference, error) {
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
	return toReferenceDomain(userID, &record), nil
}

// FindByUserID returns all references for the given user.
func (r *ReferenceRepository) FindByUserID(ctx context.Context, userID string) ([]domainreference.WorkbookReference, error) {
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
		refs = append(refs, *toReferenceDomain(userID, &record))
	}
	return refs, nil
}

// Delete removes a reference document.
func (r *ReferenceRepository) Delete(ctx context.Context, userID string, referenceID string) error {
	_, err := r.refsCol(userID).Doc(referenceID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete reference: %w", err)
	}
	return nil
}
