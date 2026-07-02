package session

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrSessionNotFound is returned when a session selector has no match.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionAmbiguous is returned when a partial selector matches multiple sessions.
var ErrSessionAmbiguous = errors.New("session selector is ambiguous")

// Ref identifies one session file.
type Ref struct {
	ID   string
	Path string
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

	path := s.path(ref)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
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

// Fork writes a new session containing the path from root to entryID.
func (s *Store) Fork(ctx context.Context, source Ref, target Ref, entryID string) error {
	sourceSession, err := s.Load(ctx, source)
	if err != nil {
		return err
	}
	entries, err := PathTo(sourceSession.Tree, entryID)
	if err != nil {
		return err
	}
	return s.createFromEntries(ctx, target, sourceSession.Header, entries)
}

// Clone writes a new session containing the active branch.
func (s *Store) Clone(ctx context.Context, source Ref, target Ref) error {
	sourceSession, err := s.Load(ctx, source)
	if err != nil {
		return err
	}
	entries, err := PathTo(sourceSession.Tree, sourceSession.Tree.ActiveLeaf())
	if err != nil {
		return err
	}
	return s.createFromEntries(ctx, target, sourceSession.Header, entries)
}

// Export writes the validated JSONL session bytes to w.
func (s *Store) Export(ctx context.Context, ref Ref, w io.Writer) error {
	if _, err := s.Load(ctx, ref); err != nil {
		return err
	}
	data, err := os.ReadFile(s.path(ref))
	if err != nil {
		return fmt.Errorf("read session export %s: %w", s.path(ref), err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write session export: %w", err)
	}
	return nil
}

// Import validates JSONL session bytes and writes them as a new session file.
func (s *Store) Import(ctx context.Context, r io.Reader, target Ref) (Ref, error) {
	if err := ctx.Err(); err != nil {
		return Ref{}, fmt.Errorf("import session: %w", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return Ref{}, fmt.Errorf("read session import: %w", err)
	}
	header, entries, err := parseSession(data)
	if err != nil {
		return Ref{}, err
	}
	if target.ID == "" {
		target.ID = header.ID
	}
	header.ID = target.ID
	if err := s.writeSession(target, header, entries); err != nil {
		return Ref{}, err
	}
	return target, nil
}

func (s *Store) path(ref Ref) string {
	if ref.Path != "" {
		return filepath.Clean(ref.Path)
	}
	return filepath.Join(s.root, ref.ID+".jsonl")
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

func (s *Store) createFromEntries(ctx context.Context, target Ref, source Header, entries []Entry) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	if target.ID == "" {
		return fmt.Errorf("create session: missing target id")
	}
	header := source
	header.ID = target.ID
	header.CreatedAt = time.Now().UTC()
	return s.writeSession(target, header, entries)
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

func readHeader(path string) (Header, error) {
	file, err := os.Open(path)
	if err != nil {
		return Header{}, fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Header{}, fmt.Errorf("read session header %s: %w", path, err)
		}
		return Header{}, fmt.Errorf("read session header %s: empty file", path)
	}
	header, _, err := parseSession(scanner.Bytes())
	if err != nil {
		return Header{}, fmt.Errorf("decode session header %s: %w", path, err)
	}
	return header, nil
}

func parseSession(data []byte) (Header, []Entry, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Header{}, nil, fmt.Errorf("read session header: %w", err)
		}
		return Header{}, nil, fmt.Errorf("read session header: empty data")
	}
	var header Header
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return Header{}, nil, fmt.Errorf("decode session header: %w", err)
	}
	if header.Type != "session" || header.Schema != schemaVersion || header.ID == "" {
		return Header{}, nil, fmt.Errorf("decode session header: invalid header")
	}
	var entries []Entry
	for scanner.Scan() {
		entry, err := UnmarshalEntry(scanner.Bytes())
		if err != nil {
			return Header{}, nil, fmt.Errorf("decode session entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return Header{}, nil, fmt.Errorf("read session: %w", err)
	}
	// Import validates tree links before any bytes are written to the store.
	if _, err := BuildTree(entries); err != nil {
		return Header{}, nil, err
	}
	return header, entries, nil
}

func writeJSONLine(file *os.File, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	return err
}
