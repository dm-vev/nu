package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const schemaVersion = 1

// ErrInvalidEntry is returned for malformed persisted session entries.
var ErrInvalidEntry = errors.New("invalid session entry")

// Kind is a persisted session entry kind.
type Kind string

const (
	KindMessage        Kind = "message"
	KindModelChange    Kind = "model_change"
	KindThinkingChange Kind = "thinking_change"
	KindLabel          Kind = "label"
	KindCompaction     Kind = "compaction"
	KindBranchSummary  Kind = "branch_summary"
	KindExtension      Kind = "extension"
)

// Entry is one non-header session JSONL record.
type Entry struct {
	Type      string          `json:"type"`
	Schema    int             `json:"schema"`
	ID        string          `json:"id"`
	ParentID  string          `json:"parent_id,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	Kind      Kind            `json:"kind"`
	Payload   json.RawMessage `json:"payload"`
}

// MarshalEntry marshals one entry without a trailing newline.
func MarshalEntry(e Entry) ([]byte, error) {
	if e.Type == "" {
		e.Type = "entry"
	}
	if e.Schema == 0 {
		e.Schema = schemaVersion
	}
	if e.Payload == nil {
		e.Payload = json.RawMessage(`{}`)
	}
	if err := validateEntry(e); err != nil {
		return nil, err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshal entry %s: %w", e.ID, err)
	}
	return data, nil
}

// UnmarshalEntry unmarshals one JSONL entry line.
func UnmarshalEntry(line []byte) (Entry, error) {
	line = bytes.TrimSuffix(line, []byte("\r"))
	var e Entry
	if err := json.Unmarshal(line, &e); err != nil {
		return Entry{}, fmt.Errorf("unmarshal entry: %w", err)
	}
	if err := validateEntry(e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

func validateEntry(e Entry) error {
	if e.Type != "entry" {
		return fmt.Errorf("%w: type must be entry", ErrInvalidEntry)
	}
	if e.Schema != schemaVersion {
		return fmt.Errorf("%w: unsupported schema %d", ErrInvalidEntry, e.Schema)
	}
	if e.ID == "" {
		return fmt.Errorf("%w: missing id", ErrInvalidEntry)
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("%w: missing created_at", ErrInvalidEntry)
	}
	if e.Kind == "" {
		return fmt.Errorf("%w: missing kind", ErrInvalidEntry)
	}
	return nil
}
