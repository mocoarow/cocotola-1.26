package gateway

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

type ownedWorkbookListRecord struct {
	WorkbookIDs []string `firestore:"workbookIDs"`
	Version     int      `firestore:"version"`
}

func (r *ownedWorkbookListRecord) GetVersion() int {
	return r.Version
}

// OwnedWorkbookListRepository manages owned workbook list persistence in Firestore.
type OwnedWorkbookListRepository struct {
	client *firestore.Client
}

// NewOwnedWorkbookListRepository returns a new OwnedWorkbookListRepository.
func NewOwnedWorkbookListRepository(client *firestore.Client) *OwnedWorkbookListRepository {
	return &OwnedWorkbookListRepository{client: client}
}

func (r *OwnedWorkbookListRepository) ownerDoc(ownerID string) *firestore.DocumentRef {
	return r.client.Collection(usersCollection).Doc(ownerID)
}

// FindByOwnerID returns the owned workbook list for the given user.
func (r *OwnedWorkbookListRepository) FindByOwnerID(ctx context.Context, ownerID string) (*domain.OwnedWorkbookList, error) {
	snap, err := r.ownerDoc(ownerID).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("get owner doc: %w", err)
		}
		// Document does not exist yet; return empty list.
		list, err := domain.NewOwnedWorkbookList(ownerID, nil)
		if err != nil {
			return nil, fmt.Errorf("new owned workbook list: %w", err)
		}
		return list, nil
	}

	var record ownedWorkbookListRecord
	if err := snap.DataTo(&record); err != nil {
		return nil, fmt.Errorf("decode owned workbook list: %w", err)
	}

	list, err := domain.NewOwnedWorkbookList(ownerID, record.WorkbookIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct owned workbook list: %w", err)
	}
	list.SetVersion(record.Version)
	return list, nil
}

// Save persists the owned workbook list atomically using a Firestore transaction.
// It uses optimistic locking via a version field on the owner document.
func (r *OwnedWorkbookListRepository) Save(ctx context.Context, list *domain.OwnedWorkbookList) error {
	record := ownedWorkbookListRecord{
		WorkbookIDs: list.Entries(),
		Version:     list.Version() + 1,
	}
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*ownedWorkbookListRecord]{
		Client:    r.client,
		Entity:    list,
		DocRef:    r.ownerDoc(list.OwnerID()),
		NewRecord: &record,
		Decode: func(snap *firestore.DocumentSnapshot) (int, error) {
			var rec ownedWorkbookListRecord
			if err := snap.DataTo(&rec); err != nil {
				return 0, fmt.Errorf("decode owned workbook list: %w", err)
			}
			return rec.Version, nil
		},
		EntityName: "owned workbook list",
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return domain.ErrOwnedWorkbookListNotFound
	}
	if err != nil {
		return fmt.Errorf("save owned workbook list: %w", err)
	}
	return nil
}
