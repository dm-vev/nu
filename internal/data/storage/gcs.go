package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	gcsstorage "cloud.google.com/go/storage"
)

// GCSStorage stores images in Google Cloud Storage.
type GCSStorage struct {
	client              *gcsstorage.Client
	bucket              string
	prefix              string
	signedURLExpiration time.Duration
	useSignedURLs       bool
}

// NewGCS creates a GCS storage backend.
func NewGCS(cfg GCSStorageConfig) (Storage, error) {
	ctx := context.Background()

	// Build client options
	opts := gcsCredentialsClientOptions(cfg)

	// Create GCS client
	client, err := gcsstorage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Validate bucket exists
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("GCS bucket name is required")
	}

	s := &GCSStorage{
		client:              client,
		bucket:              cfg.Bucket,
		prefix:              strings.TrimSuffix(cfg.Prefix, "/"),
		signedURLExpiration: cfg.SignedURLExpiration,
		useSignedURLs:       cfg.UseSignedURLs,
	}

	// Set defaults
	if s.signedURLExpiration == 0 {
		s.signedURLExpiration = 24 * time.Hour
	}

	return s, nil
}

// Name returns the storage backend name
func (s *GCSStorage) Name() string {
	return "gcs"
}
