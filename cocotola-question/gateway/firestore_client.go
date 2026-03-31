package gateway

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient creates a new Firestore client for the given project ID.
// If endpoint is non-empty, it connects to a Firestore emulator.
func NewFirestoreClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("create firestore client: %w", err)
	}
	return client, nil
}
