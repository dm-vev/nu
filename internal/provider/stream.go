package provider

import (
	"errors"
	"fmt"
	"io"
)

// EventType is a normalized provider stream event type.
type EventType string

const (
	EventStart EventType = "start"
	EventText  EventType = "text_delta"
	EventDone  EventType = "done"
	EventError EventType = "error"
)

// ErrStream is returned for provider stream failures.
var ErrStream = errors.New("provider stream error")

// Event is one normalized provider stream event.
type Event struct {
	Type       EventType
	Provider   string
	API        string
	Model      string
	Index      int
	Delta      string
	StopReason string
	ErrorClass string
	Message    string
}

// Collect drains provider events until done or error.
func Collect(ch <-chan Event) ([]Event, error) {
	var events []Event
	for ev := range ch {
		events = append(events, ev)
		switch ev.Type {
		case EventDone:
			return events, nil
		case EventError:
			return events, fmt.Errorf("%w: %s", ErrStream, ev.Message)
		}
	}
	return events, fmt.Errorf("%w: %w", ErrStream, io.ErrUnexpectedEOF)
}
