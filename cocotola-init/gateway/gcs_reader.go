package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/storage"
)

// maxObjectBytes caps how much of a CSV object is read into memory, guarding
// against a misconfigured or oversized object exhausting the process.
const maxObjectBytes = 64 << 20 // 64 MiB

// ErrObjectTooLarge is returned when an object exceeds maxObjectBytes. It is a
// sentinel so callers can distinguish "too large" from a transport error.
var ErrObjectTooLarge = errors.New("object exceeds maximum size")

// GCSReader downloads objects from a single Google Cloud Storage bucket.
// Caller owns the lifetime; call Close when finished.
type GCSReader struct {
	client *storage.Client
	bucket string
}

// NewGCSReader creates a Storage client bound to the given bucket. It uses
// Application Default Credentials, matching the rest of the platform.
func NewGCSReader(ctx context.Context, bucket string) (*GCSReader, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("new storage client: %w", err)
	}
	return &GCSReader{client: client, bucket: bucket}, nil
}

// Close releases the underlying client.
func (r *GCSReader) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("close storage client: %w", err)
	}
	return nil
}

// ReadObject returns the full contents of the object at objectKey, up to
// maxObjectBytes.
func (r *GCSReader) ReadObject(ctx context.Context, objectKey string) ([]byte, error) {
	obj := r.client.Bucket(r.bucket).Object(objectKey)
	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("open object %s: %w", objectKey, err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			slog.WarnContext(ctx, "close gcs object reader",
				slog.String("object", objectKey),
				slog.Any("error", closeErr))
		}
	}()

	data, err := ReadAllWithCap(rc, maxObjectBytes)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", objectKey, err)
	}
	return data, nil
}

// ReadAllWithCap reads all bytes from r, up to maxBytes. It reads one byte past
// the cap so an oversized source is detected and fails loudly with
// ErrObjectTooLarge instead of being silently truncated mid-row.
func ReadAllWithCap(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf(">%d bytes: %w", maxBytes, ErrObjectTooLarge)
	}
	return data, nil
}
