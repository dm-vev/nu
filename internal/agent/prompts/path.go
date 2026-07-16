package prompts

import (
	"path/filepath"
	"strings"
)

// isPathSafe checks if a file path is safe to access
func isPathSafe(filePath string, basePath string) bool {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	// Ensure path is within base directory
	return strings.HasPrefix(absPath, basePath)
}
