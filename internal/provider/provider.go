package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidRequest is returned before any provider network work starts.
var ErrInvalidRequest = errors.New("invalid provider request")

// ErrStream is returned for provider stream failures.
var ErrStream = errors.New("provider stream error")

// EventType is a normalized provider stream event type.
type EventType string

const (
	EventStart EventType = "start"
	EventText  EventType = "text_delta"
	EventDone  EventType = "done"
	EventError EventType = "error"
)

// Message is provider-neutral prompt context.
type Message struct {
	Role    string
	Content string
}

// Request is the provider-neutral request consumed by adapters.
type Request struct {
	Provider string
	API      string
	Model    string
	Messages []Message
}

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

// Streamer is the provider adapter contract consumed by the agent.
type Streamer interface {
	Stream(ctx context.Context, req Request) (<-chan Event, error)
}

// ValidateRequest validates shared request fields only.
func ValidateRequest(req Request) error {
	if req.Provider == "" {
		return fmt.Errorf("%w: missing provider", ErrInvalidRequest)
	}
	if req.API == "" {
		return fmt.Errorf("%w: missing api", ErrInvalidRequest)
	}
	if req.Model == "" {
		return fmt.Errorf("%w: missing model", ErrInvalidRequest)
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("%w: missing messages", ErrInvalidRequest)
	}
	return nil
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
