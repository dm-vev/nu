package message

import (
	"bytes"
	"errors"
	"testing"
)

func TestNUF061JSONModeEmitsSessionThenEvents(t *testing.T) {
	data, err := MarshalJSONL(Event{
		Type: "session",
		Data: map[string]string{"id": "s1"},
	})
	if err != nil {
		t.Fatalf("MarshalJSONL() error = %v", err)
	}
	if bytes.Contains(data, []byte("\n")) {
		t.Fatalf("MarshalJSONL() = %q, want no newline", data)
	}
	if !bytes.Contains(data, []byte(`"type":"session"`)) {
		t.Fatalf("MarshalJSONL() = %q, want session event", data)
	}
}

func TestMarshalJSONLRejectsEmptyType(t *testing.T) {
	_, err := MarshalJSONL(Event{})
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("MarshalJSONL() error = %v, want ErrInvalidEvent", err)
	}
}
