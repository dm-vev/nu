package gcs

import (
	"context"
	"fmt"
	"io"
	"time"

	cloudstorage "cloud.google.com/go/storage"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/storage"
)

// Store saves an image to GCS and returns an accessible URL
func (s *Storage) Store(ctx context.Context, image *contracts.GeneratedImage, metadata storage.Metadata) (string, error) {
	if image == nil || len(image.Data) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	// Build object path: prefix/orgID/threadID/timestamp_hash.ext
	objectPath := s.prefix
	if metadata.OrgID != "" {
		objectPath = joinGCSPath(objectPath, sanitizeGCSPath(metadata.OrgID))
	}
	if metadata.ThreadID != "" {
		objectPath = joinGCSPath(objectPath, sanitizeGCSPath(metadata.ThreadID))
	}

	// Generate filename: timestamp_hash.ext
	ext := getGCSExtension(image.MimeType)
	hash := hashGCSData(image.Data)[:12]
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s%s", timestamp, hash, ext)
	objectPath = joinGCSPath(objectPath, filename)

	// Get bucket handle
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(objectPath)

	// Create writer
	wc := obj.NewWriter(ctx)
	wc.ContentType = image.MimeType

	// Add metadata
	wc.Metadata = map[string]string{
		"prompt": truncateGCSString(metadata.Prompt, 500),
	}
	if metadata.OrgID != "" {
		wc.Metadata["org_id"] = metadata.OrgID
	}
	if metadata.ThreadID != "" {
		wc.Metadata["thread_id"] = metadata.ThreadID
	}
	if metadata.MessageID != "" {
		wc.Metadata["message_id"] = metadata.MessageID
	}

	// Write data
	if _, err := wc.Write(image.Data); err != nil {
		return "", fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	// Generate URL
	if s.useSignedURLs {
		return s.generateSignedURL(ctx, objectPath)
	}

	// Return public URL (requires bucket to be public or have appropriate IAM)
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, objectPath), nil
}

// Delete removes an image from GCS
func (s *Storage) Delete(ctx context.Context, url string) error {
	objectPath := s.urlToObjectPath(url)
	if objectPath == "" {
		return fmt.Errorf("invalid URL or object path")
	}

	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(objectPath)

	if err := obj.Delete(ctx); err != nil {
		if err == cloudstorage.ErrObjectNotExist {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}

// Get retrieves image data from GCS
func (s *Storage) Get(ctx context.Context, url string) ([]byte, error) {
	objectPath := s.urlToObjectPath(url)
	if objectPath == "" {
		return nil, fmt.Errorf("invalid URL or object path")
	}

	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(objectPath)

	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}
