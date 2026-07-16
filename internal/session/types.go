package session

import (
	"errors"
	"sync"
	"time"
)

// ErrSessionNotFound is returned when a session selector has no match.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionAmbiguous is returned when a partial selector matches multiple sessions.
var ErrSessionAmbiguous = errors.New("session selector is ambiguous")

// ErrImportTooLarge is returned before parsing oversized session imports.
var ErrImportTooLarge = errors.New("session import too large")

const maxImportBytes = 32 * 1024 * 1024

// Ref identifies one session file.
type Ref struct {
	ID   string
	Path string
	CWD  string
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
