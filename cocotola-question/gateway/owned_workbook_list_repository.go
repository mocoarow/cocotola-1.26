package gateway

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

type ownedWorkbookListRecord struct {
	WorkbookIDs []string `firestore:"workbookIDs"`
	Version     int      `firestore:"version"`
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
	nextVersion := list.Version() + 1
	if err := r.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		ownerRef := r.ownerDoc(list.OwnerID())

		// Verify version (optimistic lock).
		snap, err := tx.Get(ownerRef)
		currentVersion := 0
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("get owner doc in tx: %w", err)
			}
		} else {
			var record ownedWorkbookListRecord
			if err := snap.DataTo(&record); err != nil {
				return fmt.Errorf("decode owner doc in tx: %w", err)
			}
			currentVersion = record.Version
		}

		if currentVersion != list.Version() {
			return domain.ErrConcurrentModification
		}

		record := ownedWorkbookListRecord{
			WorkbookIDs: list.Entries(),
			Version:     nextVersion,
		}
		if err := tx.Set(ownerRef, record); err != nil {
			return fmt.Errorf("save owned workbook list: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("run transaction: %w", err)
	}

	list.SetVersion(nextVersion)
	return nil
}
