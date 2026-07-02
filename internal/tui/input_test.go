package tui

import "testing"

func TestNUF100InputDecodesChunkedEscape(t *testing.T) {
	decoder := NewDecoder()

	if events := decoder.Write([]byte("\x1b")); len(events) != 0 {
		t.Fatalf("events after partial escape = %#v, want none", events)
	}
	events := decoder.Write([]byte("[A"))

	if len(events) != 1 || events[0].Kind != EventKey || events[0].Key != "up" {
		t.Fatalf("events = %#v, want up key", events)
	}
}

func TestNUF100BracketedPasteIsSingleEvent(t *testing.T) {
	decoder := NewDecoder()

	events := decoder.Write([]byte("\x1b[200~hello\nworld\x1b[201~"))

	if len(events) != 1 {
		t.Fatalf("events = %#v, want one paste event", events)
	}
	if events[0].Kind != EventPaste || events[0].Text != "hello\nworld" {
		t.Fatalf("paste event = %#v", events[0])
	}
}
