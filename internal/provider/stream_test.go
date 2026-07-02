package provider

import (
	"errors"
	"testing"
)

func TestProviderCollectStopsAtDone(t *testing.T) {
	ch := make(chan Event, 3)
	ch <- Event{Type: EventStart}
	ch <- Event{Type: EventDone}
	ch <- Event{Type: EventText, Delta: "ignored"}
	close(ch)

	events, err := Collect(ch)
	if err != nil {
		t.Fatalf("Collect error = %v, want nil", err)
	}
	if len(events) != 2 {
		t.Fatalf("Collect events = %d, want 2", len(events))
	}
}

func TestProviderCollectRejectsErrorEvent(t *testing.T) {
	ch := make(chan Event, 3)
	ch <- Event{Type: EventStart}
	ch <- Event{Type: EventError, Message: "boom"}
	ch <- Event{Type: EventText, Delta: "ignored"}
	close(ch)

	events, err := Collect(ch)
	if !errors.Is(err, ErrStream) {
		t.Fatalf("Collect error = %v, want ErrStream", err)
	}
	if len(events) != 2 {
		t.Fatalf("Collect events = %d, want 2", len(events))
	}
}

func TestProviderCollectRejectsUnexpectedEOF(t *testing.T) {
	ch := make(chan Event, 1)
	ch <- Event{Type: EventStart}
	close(ch)

	events, err := Collect(ch)
	if !errors.Is(err, ErrStream) {
		t.Fatalf("Collect error = %v, want ErrStream", err)
	}
	if len(events) != 1 {
		t.Fatalf("Collect events = %d, want 1", len(events))
	}
}
