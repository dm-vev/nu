package testkit

import (
	"context"
	"fmt"
	"sync"

	"nu/internal/provider"
)

// ScriptedProvider emits a fixed event script for tests.
type ScriptedProvider struct {
	mu       sync.Mutex
	scripts  [][]provider.Event
	errors   []error
	requests []provider.Request
}

// NewScriptedProvider creates a deterministic fake provider.
func NewScriptedProvider(events ...provider.Event) *ScriptedProvider {
	return NewScriptedProviderScripts(events)
}

// NewScriptedProviderScripts creates a fake provider with one script per request.
func NewScriptedProviderScripts(scripts ...[]provider.Event) *ScriptedProvider {
	copied := make([][]provider.Event, 0, len(scripts))
	for _, script := range scripts {
		copied = append(copied, append([]provider.Event(nil), script...))
	}
	return &ScriptedProvider{scripts: copied}
}

// NewScriptedProviderErrors creates a fake provider with one start result per request.
func NewScriptedProviderErrors(errors []error, scripts ...[]provider.Event) *ScriptedProvider {
	fake := NewScriptedProviderScripts(scripts...)
	fake.errors = append([]error(nil), errors...)
	return fake
}

// Stream records req and emits scripted events.
func (p *ScriptedProvider) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error) {
	p.mu.Lock()
	index := len(p.requests)
	p.requests = append(p.requests, req)
	if index < len(p.errors) && p.errors[index] != nil {
		err := p.errors[index]
		p.mu.Unlock()
		return nil, err
	}
	scriptIndex := index - p.errorCountBefore(index)
	if scriptIndex >= len(p.scripts) {
		p.mu.Unlock()
		return nil, fmt.Errorf("scripted provider: missing script %d", scriptIndex)
	}
	events := append([]provider.Event(nil), p.scripts[scriptIndex]...)
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

func (p *ScriptedProvider) errorCountBefore(index int) int {
	count := 0
	for i := 0; i < index && i < len(p.errors); i++ {
		if p.errors[i] != nil {
			count++
		}
	}
	return count
}

// Requests returns recorded provider requests.
func (p *ScriptedProvider) Requests() []provider.Request {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]provider.Request(nil), p.requests...)
}
