package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
)

// generateSignedURL creates a signed URL for the object
func (s *GCSStorage) generateSignedURL(ctx context.Context, objectPath string) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(s.signedURLExpiration),
	}

	url, err := s.client.Bucket(s.bucket).SignedURL(objectPath, opts)
	if err != nil {
		// Fall back to public URL if signing fails
		return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, objectPath), nil
	}

	return url, nil
}
