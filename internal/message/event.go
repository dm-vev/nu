package message

import (
	"encoding/json"
	"errors"
)

// ErrInvalidEvent is returned when an event cannot be represented on JSONL.
var ErrInvalidEvent = errors.New("invalid event")

// Event is one runtime event emitted to JSON/RPC consumers.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// MarshalJSONL marshals one event object without the JSONL newline.
func MarshalJSONL(event Event) ([]byte, error) {
	if event.Type == "" {
		return nil, ErrInvalidEvent
	}
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return data, nil
}
