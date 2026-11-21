package gateway

import (
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// collectDocs drains iter, converting each document with decode, and stops
// iter before returning. It stops at the first error.
func collectDocs[D any](iter *firestore.DocumentIterator, decode func(*firestore.DocumentSnapshot) (D, error)) ([]D, error) {
	defer iter.Stop()

	var items []D

	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return items, nil
		}
		if err != nil {
			return nil, fmt.Errorf("next document: %w", err)
		}
		item, err := decode(doc)
		if err != nil {
			return nil, fmt.Errorf("document %s: %w", doc.Ref.ID, err)
		}
		items = append(items, item)
	}
}
