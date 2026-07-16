package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Resolve resolves a path, full id, or partial id to a session ref.
func (s *Store) Resolve(ctx context.Context, selector string) (Ref, error) {
	if err := ctx.Err(); err != nil {
		return Ref{}, fmt.Errorf("resolve session: %w", err)
	}
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return Ref{}, fmt.Errorf("%w: empty selector", ErrSessionNotFound)
	}
	// Existing JSONL paths bypass root lookup so explicit resume paths work.
	if ref, ok, err := s.resolvePath(selector); ok || err != nil {
		return ref, err
	}

	refs, err := s.matchRefs(selector)
	if err != nil {
		return Ref{}, err
	}
	switch len(refs) {
	case 0:
		return Ref{}, fmt.Errorf("%w: %s", ErrSessionNotFound, selector)
	case 1:
		return refs[0], nil
	default:
		return Ref{}, fmt.Errorf("%w: %s", ErrSessionAmbiguous, selector)
	}
}

// LatestByCWD returns the newest session whose header cwd matches cwd.
func (s *Store) LatestByCWD(ctx context.Context, cwd string) (Ref, error) {
	if err := ctx.Err(); err != nil {
		return Ref{}, fmt.Errorf("latest session: %w", err)
	}
	cwd = filepath.Clean(cwd)
	refs, err := s.sessionRefs()
	if err != nil {
		return Ref{}, err
	}
	var best Ref
	var bestCreated time.Time
	for _, ref := range refs {
		// Header-only reads keep continue lookup cheap even with large sessions.
		header, err := readHeader(s.path(ref))
		if err != nil {
			return Ref{}, err
		}
		if filepath.Clean(header.CWD) != cwd {
			continue
		}
		if best.ID == "" || header.CreatedAt.After(bestCreated) || header.CreatedAt.Equal(bestCreated) && ref.ID > best.ID {
			best = ref
			bestCreated = header.CreatedAt
		}
	}
	if best.ID == "" {
		return Ref{}, fmt.Errorf("%w: cwd %s", ErrSessionNotFound, cwd)
	}
	return best, nil
}

func (s *Store) resolvePath(selector string) (Ref, bool, error) {
	clean := filepath.Clean(selector)
	if !strings.HasSuffix(clean, ".jsonl") && !strings.Contains(clean, string(os.PathSeparator)) {
		return Ref{}, false, nil
	}
	info, err := os.Stat(clean)
	if errors.Is(err, os.ErrNotExist) {
		return Ref{}, false, nil
	}
	if err != nil {
		return Ref{}, true, fmt.Errorf("stat session %s: %w", clean, err)
	}
	if info.IsDir() {
		return Ref{}, true, fmt.Errorf("%w: %s is a directory", ErrSessionNotFound, clean)
	}
	header, err := readHeader(clean)
	if err != nil {
		return Ref{}, true, err
	}
	return Ref{ID: header.ID, Path: clean}, true, nil
}

func (s *Store) matchRefs(selector string) ([]Ref, error) {
	refs, err := s.sessionRefs()
	if err != nil {
		return nil, err
	}
	var matches []Ref
	for _, ref := range refs {
		if ref.ID == selector || strings.HasPrefix(ref.ID, selector) {
			matches = append(matches, ref)
		}
	}
	return matches, nil
}

func (s *Store) sessionRefs() ([]Ref, error) {
	entries, err := os.ReadDir(s.root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read session root %s: %w", s.root, err)
	}
	refs := make([]Ref, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".jsonl")
		refs = append(refs, Ref{ID: id})
	}
	return refs, nil
}
