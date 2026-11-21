package gateway

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
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

// Save persists a workbook reference. Create on a workbookID-keyed doc enforces
// uniqueness atomically, avoiding a read-then-write race.
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
	ref, err := decodeReferenceDoc(doc, userID)
	if err != nil {
		return nil, fmt.Errorf("find reference %s: %w", referenceID, err)
	}
	return &ref, nil
}

func decodeReferenceDoc(doc *firestore.DocumentSnapshot, userID string) (domainreference.WorkbookReference, error) {
	var record referenceRecord
	if err := doc.DataTo(&record); err != nil {
		return domainreference.WorkbookReference{}, fmt.Errorf("decode reference: %w", err)
	}
	return *toReferenceDomain(userID, &record), nil
}

// FindByUserID returns all references for the given user.
func (r *ReferenceRepository) FindByUserID(ctx context.Context, userID string) ([]domainreference.WorkbookReference, error) {
	iter := r.refsCol(userID).Documents(ctx)
	refs, err := collectDocs(iter, func(doc *firestore.DocumentSnapshot) (domainreference.WorkbookReference, error) {
		return decodeReferenceDoc(doc, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("find references by user %s: %w", userID, err)
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
