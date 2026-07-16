package coding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveUnder(cwd, requested string) (string, string, error) {
	if strings.TrimSpace(cwd) == "" {
		return "", "", errors.New("missing cwd")
	}
	if strings.TrimSpace(requested) == "" {
		return "", "", errors.New("missing path")
	}
	if filepath.IsAbs(requested) {
		return "", "", fmt.Errorf("path %q must be relative", requested)
	}
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return "", "", fmt.Errorf("resolve cwd %q: %w", cwd, err)
	}
	realCWD, err := filepath.EvalSymlinks(absCWD)
	if err != nil {
		return "", "", fmt.Errorf("resolve cwd %q: %w", cwd, err)
	}
	clean := filepath.Clean(requested)
	if clean == "." {
		return realCWD, ".", nil
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path %q escapes cwd", requested)
	}
	path := filepath.Join(realCWD, clean)
	realPath, err := resolveForContainment(path)
	if err != nil {
		return "", "", err
	}
	if !pathUnder(realCWD, realPath) {
		return "", "", fmt.Errorf("path %q escapes cwd", requested)
	}
	return realPath, filepath.ToSlash(clean), nil
}

func resolveForContainment(path string) (string, error) {
	var missing []string
	current := path
	for {
		real, err := filepath.EvalSymlinks(current)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				real = filepath.Join(real, missing[i])
			}
			return real, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve path %q: %w", path, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("resolve path %q: %w", path, err)
		}
		// Missing leaf components are allowed for write/create tools, but the nearest
		// existing parent must still resolve under cwd.
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func pathUnder(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func relative(cwd, path string) (string, error) {
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	return filepath.ToSlash(rel), nil
}

func defaultPath(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
