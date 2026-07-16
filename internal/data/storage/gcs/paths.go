package gcs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// urlToObjectPath extracts the object path from a URL
func (s *Storage) urlToObjectPath(url string) string {
	// Handle direct object paths
	if !strings.HasPrefix(url, "http") {
		return url
	}

	// Handle GCS URLs: https://storage.googleapis.com/bucket/path
	prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.bucket)
	if strings.HasPrefix(url, prefix) {
		path := strings.TrimPrefix(url, prefix)
		// Remove query parameters (for signed URLs)
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}
		return path
	}

	// Handle signed URLs with the bucket in the path
	if strings.Contains(url, s.bucket) {
		// Extract path after bucket name
		parts := strings.SplitN(url, s.bucket+"/", 2)
		if len(parts) == 2 {
			path := parts[1]
			// Remove query parameters
			if idx := strings.Index(path, "?"); idx != -1 {
				path = path[:idx]
			}
			return path
		}
	}

	return ""
}

// getExtension returns the file extension for a MIME type
func getGCSExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

// hashData returns a SHA256 hash of the data
func hashGCSData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// sanitizePath removes potentially dangerous characters from path components
func sanitizeGCSPath(s string) string {
	s = strings.ReplaceAll(s, "..", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	return s
}

// joinPath joins path components with forward slashes
func joinGCSPath(base, path string) string {
	if base == "" {
		return path
	}
	if path == "" {
		return base
	}
	return base + "/" + path
}

// truncateString truncates a string to maxLen characters
func truncateGCSString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
