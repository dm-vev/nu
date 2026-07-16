package gcs

import (
	"context"
	"fmt"
	"strings"
	"time"

	gcsstorage "cloud.google.com/go/storage"

	"github.com/dm-vev/nu/internal/data/storage"
)

// Config configures Google Cloud Storage.
type Config struct {
	Bucket              string
	Prefix              string
	CredentialsFile     string
	CredentialsJSON     string
	SignedURLExpiration time.Duration
	UseSignedURLs       bool
}

// Storage stores images in Google Cloud Storage.
type Storage struct {
	client              *gcsstorage.Client
	bucket              string
	prefix              string
	signedURLExpiration time.Duration
	useSignedURLs       bool
}

// New creates a GCS storage backend.
func New(cfg Config) (storage.Storage, error) {
	ctx := context.Background()

	// Build client options
	opts := credentialsClientOptions(cfg)

	// Create GCS client
	client, err := gcsstorage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Validate bucket exists
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("GCS bucket name is required")
	}

	s := &Storage{
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
func (s *Storage) Name() string {
	return "gcs"
}
