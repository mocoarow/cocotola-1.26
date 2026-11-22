package gateway

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/firestore"
)

// InitFirestore creates a Firestore client for the given project ID and
// returns a function that closes it.
func InitFirestore(ctx context.Context, projectID string) (*firestore.Client, func(), error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("create firestore client: %w", err)
	}
	return client, func() {
		if err := client.Close(); err != nil {
			slog.Error("close firestore client", slog.Any("error", err))
		}
	}, nil
}
