package testkit

import (
	"context"
	"sync"

	"nu/internal/provider"
)

// ScriptedProvider emits a fixed event script for tests.
type ScriptedProvider struct {
	mu       sync.Mutex
	events   []provider.Event
	requests []provider.Request
}

// NewScriptedProvider creates a deterministic fake provider.
func NewScriptedProvider(events ...provider.Event) *ScriptedProvider {
	copied := append([]provider.Event(nil), events...)
	return &ScriptedProvider{events: copied}
}

// Stream records req and emits scripted events.
func (p *ScriptedProvider) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error) {
	p.mu.Lock()
	p.requests = append(p.requests, req)
	events := append([]provider.Event(nil), p.events...)
	p.mu.Unlock()

	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		for _, ev := range events {
			select {
			case <-ctx.Done():
				return
			case ch <- ev:
			}
		}
	}()
	return ch, nil
}

// Requests returns recorded provider requests.
func (p *ScriptedProvider) Requests() []provider.Request {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]provider.Request(nil), p.requests...)
}
