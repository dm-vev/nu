package prompts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PromptFileStore implements PromptStore using the local file system.
type PromptFileStore struct {
	basePath string
}

// NewPromptFileStore creates a file-backed prompt store.
func NewPromptFileStore(basePath string) (*PromptFileStore, error) {
	// Create directory if it doesn't exist
	err := os.MkdirAll(basePath, 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &PromptFileStore{
		basePath: basePath,
	}, nil
}

// Get retrieves a template by ID and version
func (s *PromptFileStore) Get(ctx context.Context, id string, version string) (*PromptTemplate, error) {
	// Sanitize id and version to prevent path traversal
	id = filepath.Base(id)
	version = filepath.Base(version)

	// Construct file path
	filePath := filepath.Join(s.basePath, fmt.Sprintf("%s_%s.tmpl", id, version))

	// Ensure the file is within the basePath
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute base path: %w", err)
	}

	if !isPathSafe(filePath, absBasePath) {
		return nil, fmt.Errorf("invalid template path")
	}

	// Read file
	data, err := os.ReadFile(filePath) // #nosec G304 - Path is validated with isPathSafe() before use
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template
	tmpl, err := parseTemplateFile(string(data), id, version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template file: %w", err)
	}

	return tmpl, nil
}

// List returns all templates matching the given filter
func (s *PromptFileStore) List(ctx context.Context, filter map[string]interface{}) ([]*PromptTemplate, error) {
	// Get all template files
	pattern := filepath.Join(s.basePath, "*.tmpl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list template files: %w", err)
	}

	// Get absolute base path for validation
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute base path: %w", err)
	}

	// Parse each file
	var templates []*PromptTemplate
	for _, file := range files {
		// Verify file is within basePath
		if !isPathSafe(file, absBasePath) {
			continue
		}

		// Extract ID and version from filename
		filename := filepath.Base(file)
		parts := strings.Split(strings.TrimSuffix(filename, ".tmpl"), "_")
		if len(parts) != 2 {
			continue
		}

		id := parts[0]
		version := parts[1]

		// Read file
		data, err := os.ReadFile(file) // #nosec G304 - Path is validated with isPathSafe() before use
		if err != nil {
			continue
		}

		// Parse template
		tmpl, err := parseTemplateFile(string(data), id, version)
		if err != nil {
			continue
		}

		// Apply filter
		if matchesFilter(tmpl, filter) {
			templates = append(templates, tmpl)
		}
	}

	return templates, nil
}

// Save stores a template
func (s *PromptFileStore) Save(ctx context.Context, tmpl *PromptTemplate) error {
	// Sanitize id and version to prevent path traversal
	tmpl.ID = filepath.Base(tmpl.ID)
	tmpl.Version = filepath.Base(tmpl.Version)

	// Update timestamp
	tmpl.UpdatedAt = time.Now()

	// Serialize template
	data := serializeTemplate(tmpl)

	// Construct file path
	filePath := filepath.Join(s.basePath, fmt.Sprintf("%s_%s.tmpl", tmpl.ID, tmpl.Version))

	// Ensure the file is within the basePath
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute base path: %w", err)
	}

	if !isPathSafe(filePath, absBasePath) {
		return fmt.Errorf("invalid template path")
	}

	// Write file with secure permissions
	err = os.WriteFile(filePath, []byte(data), 0600)
	if err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// Delete removes a template
func (s *PromptFileStore) Delete(ctx context.Context, id string, version string) error {
	// Sanitize id and version to prevent path traversal
	id = filepath.Base(id)
	version = filepath.Base(version)

	// Construct file path
	filePath := filepath.Join(s.basePath, fmt.Sprintf("%s_%s.tmpl", id, version))

	// Ensure the file is within the basePath
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute base path: %w", err)
	}

	if !isPathSafe(filePath, absBasePath) {
		return fmt.Errorf("invalid template path")
	}

	// Delete file
	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	return nil
}

// matchesFilter checks if a template matches the given filter
func matchesFilter(tmpl *PromptTemplate, filter map[string]interface{}) bool {
	for key, value := range filter {
		switch key {
		case "id":
			if tmpl.ID != value {
				return false
			}
		case "name":
			if tmpl.Name != value {
				return false
			}
		case "version":
			if tmpl.Version != value {
				return false
			}
		case "tag":
			found := false
			for _, tag := range tmpl.Tags {
				if tag == value {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		default:
			metaValue, ok := tmpl.Metadata[key]
			if !ok || metaValue != value {
				return false
			}
		}
	}

	return true
}
