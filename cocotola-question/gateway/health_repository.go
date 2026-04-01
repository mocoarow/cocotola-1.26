package gateway

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// HealthRepository checks Firestore connectivity.
type HealthRepository struct {
	client *firestore.Client
}

// NewHealthRepository returns a new HealthRepository.
func NewHealthRepository(client *firestore.Client) *HealthRepository {
	return &HealthRepository{client: client}
}

// Check verifies Firestore connectivity by reading the collections list.
func (r *HealthRepository) Check(ctx context.Context) error {
	iter := r.client.Collections(ctx)
	_, err := iter.Next()
	// iterator.Done means no collections, which is fine for health check
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return nil
		}
		return fmt.Errorf("firestore health check: %w", err)
	}
	return nil
}
