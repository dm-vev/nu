package storage

import (
	"context"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// Storage defines storage for generated images.
type Storage interface {
	// Store saves an image and returns an accessible URL
	Store(ctx context.Context, image *contracts.GeneratedImage, metadata Metadata) (string, error)

	// Delete removes an image by URL
	Delete(ctx context.Context, url string) error

	// Get retrieves image data by URL (optional, may not be supported by all backends)
	Get(ctx context.Context, url string) ([]byte, error)

	// Name returns the storage backend name
	Name() string
}

// Metadata contains metadata for stored images
type Metadata struct {
	// OrgID is the organization ID for multi-tenancy
	OrgID string

	// ThreadID is the conversation thread ID
	ThreadID string

	// MessageID is the message ID
	MessageID string

	// Prompt is the original prompt used to generate the image
	Prompt string

	// Tags contains custom tags for the image
	Tags map[string]string

	// CreatedAt is the timestamp when the image was created
	CreatedAt time.Time
}
