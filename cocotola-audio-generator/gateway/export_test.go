package gateway

import "cloud.google.com/go/storage"

// NewGCSClientForTest creates a GCSClient around the given Storage client.
func NewGCSClientForTest(client *storage.Client, bucket string) *GCSClient {
	return &GCSClient{client: client, bucket: bucket}
}
