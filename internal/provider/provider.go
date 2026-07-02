package provider

import (
	"context"
	"errors"
)

// ErrUnsupported is returned when a provider/api pair is not registered.
var ErrUnsupported = errors.New("unsupported provider")

// Streamer is the provider adapter contract consumed by the agent.
type Streamer interface {
	Stream(ctx context.Context, req Request) (<-chan Event, error)
}

// Registry resolves provider streamers.
type Registry struct {
	streamers map[string]Streamer
}

// NewRegistry creates an immutable provider registry.
func NewRegistry(streamers map[string]Streamer) *Registry {
	copied := make(map[string]Streamer, len(streamers))
	for key, streamer := range streamers {
		copied[key] = streamer
	}
	return &Registry{streamers: copied}
}

// Resolve returns the streamer for provider/api.
func (r *Registry) Resolve(providerID, api string) (Streamer, error) {
	if r == nil {
		return nil, ErrUnsupported
	}
	streamer, ok := r.streamers[providerID+"/"+api]
	if !ok {
		return nil, ErrUnsupported
	}
	return streamer, nil
}
