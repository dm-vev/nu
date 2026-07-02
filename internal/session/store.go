package session

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FS is reserved for later filesystem injection.
type FS interface{}

// Ref identifies one session file.
type Ref struct {
	ID string
}

// Store manages append-only session JSONL files.
type Store struct {
	root string
	mu   sync.Mutex
}

// Session is a loaded session file.
type Session struct {
	Header  Header
	Entries []Entry
	Tree    *Tree
}

// Header is the first line in a session JSONL file.
type Header struct {
	Type       string    `json:"type"`
	Schema     int       `json:"schema"`
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	CWD        string    `json:"cwd"`
	App        string    `json:"app"`
	AppVersion string    `json:"app_version"`
}

// OpenStore creates a session store rooted at root.
func OpenStore(root string, _ FS) *Store {
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
	if !isNew && entry.ParentID != "" {
		if err := s.parentExists(path, entry.ParentID); err != nil {
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
			CWD:        "",
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

func (s *Store) parentExists(path, parentID string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()

	// Scan existing entries only until the parent is found; no full tree rebuild here.
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return fmt.Errorf("%w: missing parent %s", ErrInvalidTree, parentID)
	}
	for scanner.Scan() {
		entry, err := UnmarshalEntry(scanner.Bytes())
		if err != nil {
			return fmt.Errorf("decode session entry %s: %w", path, err)
		}
		if entry.ID == parentID {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read session %s: %w", path, err)
	}
	return fmt.Errorf("%w: missing parent %s", ErrInvalidTree, parentID)
}

// Load reconstructs a session from disk.
func (s *Store) Load(ctx context.Context, ref Ref) (*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}
	if ref.ID == "" {
		return nil, fmt.Errorf("load session: missing ref id")
	}

	path := s.path(ref)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read session header %s: %w", path, err)
		}
		return nil, fmt.Errorf("read session header %s: empty file", path)
	}

	var header Header
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return nil, fmt.Errorf("decode session header %s: %w", path, err)
	}
	if header.Type != "session" || header.Schema != schemaVersion || header.ID == "" {
		return nil, fmt.Errorf("decode session header %s: invalid header", path)
	}

	var entries []Entry
	for scanner.Scan() {
		entry, err := UnmarshalEntry(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("decode session entry %s: %w", path, err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read session %s: %w", path, err)
	}

	tree, err := BuildTree(entries)
	if err != nil {
		return nil, err
	}
	return &Session{Header: header, Entries: entries, Tree: tree}, nil
}

func (s *Store) path(ref Ref) string {
	return filepath.Join(s.root, ref.ID+".jsonl")
}

func writeJSONLine(file *os.File, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	return err
}
