package session

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OpenStore creates a session store rooted at root.
func OpenStore(root string) *Store {
	return &Store{root: filepath.Clean(root)}
}

// Append appends one entry to a session file.
func (s *Store) Append(ctx context.Context, ref Ref, entry Entry) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("append session: %w", err)
	}
	if ref.ID == "" {
		return fmt.Errorf("append session: missing ref id")
	}
	if _, err := MarshalEntry(entry); err != nil {
		return err
	}

	// Serialize append with parent validation so a bad child is never persisted.
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("create session root %s: %w", s.root, err)
	}

	path := s.path(ref)
	_, statErr := os.Stat(path)
	isNew := os.IsNotExist(statErr)
	if !isNew {
		if err := s.validateAppend(path, entry); err != nil {
			return err
		}
	}
	if isNew && entry.ParentID != "" {
		return fmt.Errorf("%w: missing parent %s", ErrInvalidTree, entry.ParentID)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()

	if isNew {
		header := Header{
			Type:       "session",
			Schema:     schemaVersion,
			ID:         ref.ID,
			CreatedAt:  time.Now().UTC(),
			CWD:        cleanOptionalPath(ref.CWD),
			App:        "nu",
			AppVersion: "0.0.0",
		}
		if err := writeJSONLine(file, header); err != nil {
			return fmt.Errorf("write session header %s: %w", path, err)
		}
	}

	data, err := MarshalEntry(entry)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("append session entry %s: %w", path, err)
	}
	return nil
}

func (s *Store) validateAppend(path string, next Entry) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()

	// Scan existing entries only until the parent is found; no full tree rebuild here.
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	if !scanner.Scan() {
		return fmt.Errorf("%w: missing parent %s", ErrInvalidTree, next.ParentID)
	}
	parentFound := next.ParentID == ""
	for scanner.Scan() {
		entry, err := UnmarshalEntry(scanner.Bytes())
		if err != nil {
			return fmt.Errorf("decode session entry %s: %w", path, err)
		}
		if entry.ID == next.ID {
			return fmt.Errorf("%w: duplicate id %s", ErrInvalidTree, next.ID)
		}
		if entry.ID == next.ParentID {
			parentFound = true
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read session %s: %w", path, err)
	}
	if !parentFound {
		return fmt.Errorf("%w: missing parent %s", ErrInvalidTree, next.ParentID)
	}
	return nil
}

// Load reconstructs a session from disk.
func (s *Store) Load(ctx context.Context, ref Ref) (*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}
	if ref.ID == "" {
		return nil, fmt.Errorf("load session: missing ref id")
	}

	header, entries, err := readSession(s.path(ref))
	if err != nil {
		return nil, err
	}
	tree, err := BuildTree(entries)
	if err != nil {
		return nil, err
	}
	return &Session{Header: header, Entries: entries, Tree: tree}, nil
}

func (s *Store) path(ref Ref) string {
	if ref.Path != "" {
		return filepath.Clean(ref.Path)
	}
	return filepath.Join(s.root, ref.ID+".jsonl")
}

func cleanOptionalPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.Clean(path)
}

func (s *Store) writeSession(ref Ref, header Header, entries []Entry) error {
	if ref.ID == "" {
		return fmt.Errorf("write session: missing ref id")
	}
	// Validate the copied/imported branch before creating the target file.
	if _, err := BuildTree(entries); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path(ref)), 0o755); err != nil {
		return fmt.Errorf("create session dir %s: %w", filepath.Dir(s.path(ref)), err)
	}
	file, err := os.OpenFile(s.path(ref), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create session %s: %w", s.path(ref), err)
	}
	defer file.Close()
	if err := writeJSONLine(file, header); err != nil {
		return fmt.Errorf("write session header %s: %w", s.path(ref), err)
	}
	for _, entry := range entries {
		data, err := MarshalEntry(entry)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write session entry %s: %w", s.path(ref), err)
		}
	}
	return nil
}
