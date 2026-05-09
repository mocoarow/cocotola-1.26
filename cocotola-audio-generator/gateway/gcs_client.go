package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/storage"
)

// GCSClient uploads audio files to a single bucket and applies the cache
// headers required for long-lived static delivery.
type GCSClient struct {
	client *storage.Client
	bucket string
}

// NewGCSClient creates a Storage client connected to the given bucket.
// Caller owns the lifetime; call Close.
func NewGCSClient(ctx context.Context, bucket string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("new storage client: %w", err)
	}
	return &GCSClient{client: client, bucket: bucket}, nil
}

// Close releases the underlying HTTP client.
func (c *GCSClient) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("close storage client: %w", err)
	}
	return nil
}

// Upload writes data to the given object path with the supplied content type
// and a long, immutable Cache-Control header so the public CDN/browser cache
// it forever (we use a different path when the underlying text changes).
//
// Returns the number of bytes written.
func (c *GCSClient) Upload(ctx context.Context, objectPath, contentType string, data []byte) (int64, error) {
	bucket := c.client.Bucket(c.bucket)
	obj := bucket.Object(objectPath)
	w := obj.NewWriter(ctx)
	w.ContentType = contentType
	w.CacheControl = "public, max-age=31536000, immutable"
	n, err := io.Copy(w, bytes.NewReader(data))
	if err != nil {
		// Best-effort cleanup. Surface the close failure separately so the
		// underlying network/IO failure remains traceable in logs.
		if closeErr := w.Close(); closeErr != nil {
			slog.WarnContext(ctx, "close object writer after copy failure",
				slog.String("objectPath", objectPath),
				slog.Any("error", closeErr))
		}
		return 0, fmt.Errorf("copy to object %s: %w", objectPath, err)
	}
	if err := w.Close(); err != nil {
		return 0, fmt.Errorf("close object %s: %w", objectPath, err)
	}
	return n, nil
}
