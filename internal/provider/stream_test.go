package provider

import (
	"errors"
	"testing"
)

func TestProviderCollectStopsOnError(t *testing.T) {
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
