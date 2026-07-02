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
	Type      string                     `json:"type"`
	Schema    int                        `json:"schema"`
	ID        string                     `json:"id"`
	ParentID  string                     `json:"parent_id,omitempty"`
	CreatedAt time.Time                  `json:"created_at"`
	Kind      Kind                       `json:"kind"`
	Payload   json.RawMessage            `json:"payload"`
	Extra     map[string]json.RawMessage `json:"-"`
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
	data, err := marshalEntryObject(e)
	if err != nil {
		return nil, fmt.Errorf("marshal entry %s: %w", e.ID, err)
	}
	return data, nil
}

// UnmarshalEntry unmarshals one JSONL entry line.
func UnmarshalEntry(line []byte) (Entry, error) {
	line = bytes.TrimSuffix(line, []byte("\r"))
	e, err := unmarshalEntryObject(line)
	if err != nil {
		return Entry{}, fmt.Errorf("unmarshal entry: %w", err)
	}
	if err := validateEntry(e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

func marshalEntryObject(e Entry) ([]byte, error) {
	base := struct {
		Type      string          `json:"type"`
		Schema    int             `json:"schema"`
		ID        string          `json:"id"`
		ParentID  string          `json:"parent_id,omitempty"`
		CreatedAt time.Time       `json:"created_at"`
		Kind      Kind            `json:"kind"`
		Payload   json.RawMessage `json:"payload"`
	}{
		Type:      e.Type,
		Schema:    e.Schema,
		ID:        e.ID,
		ParentID:  e.ParentID,
		CreatedAt: e.CreatedAt,
		Kind:      e.Kind,
		Payload:   e.Payload,
	}
	data, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	if len(e.Extra) == 0 {
		return data, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	// Extra fields are appended after known fields without letting imports override the envelope.
	for name, value := range e.Extra {
		if reservedEntryField(name) {
			continue
		}
		object[name] = value
	}
	return json.Marshal(object)
}

func unmarshalEntryObject(line []byte) (Entry, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return Entry{}, err
	}
	var e struct {
		Type      string          `json:"type"`
		Schema    int             `json:"schema"`
		ID        string          `json:"id"`
		ParentID  string          `json:"parent_id,omitempty"`
		CreatedAt time.Time       `json:"created_at"`
		Kind      Kind            `json:"kind"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(line, &e); err != nil {
		return Entry{}, err
	}
	entry := Entry{
		Type:      e.Type,
		Schema:    e.Schema,
		ID:        e.ID,
		ParentID:  e.ParentID,
		CreatedAt: e.CreatedAt,
		Kind:      e.Kind,
		Payload:   e.Payload,
	}
	// Unknown top-level fields stay available for future import/migration code.
	for name, value := range raw {
		if reservedEntryField(name) {
			continue
		}
		if entry.Extra == nil {
			entry.Extra = map[string]json.RawMessage{}
		}
		entry.Extra[name] = value
	}
	return entry, nil
}

func reservedEntryField(name string) bool {
	switch name {
	case "type", "schema", "id", "parent_id", "created_at", "kind", "payload":
		return true
	default:
		return false
	}
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
