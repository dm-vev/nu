package provider

import (
	"errors"
	"fmt"
)

// ErrInvalidRequest is returned before any provider network work starts.
var ErrInvalidRequest = errors.New("invalid provider request")

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
