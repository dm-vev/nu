package testkit

import (
	"context"
	"sync"
	"time"

	"github.com/dm-vev/nu/contracts"
)

type ScriptedAgent struct {
	mu      sync.Mutex
	scripts [][]contracts.AgentStreamEvent
	prompts []string
}

func NewScriptedAgent(events ...contracts.AgentStreamEvent) *ScriptedAgent {
	return NewScriptedAgentScripts(events)
}

func NewScriptedAgentScripts(scripts ...[]contracts.AgentStreamEvent) *ScriptedAgent {
	copyOfScripts := make([][]contracts.AgentStreamEvent, len(scripts))
	for index := range scripts {
		copyOfScripts[index] = append([]contracts.AgentStreamEvent(nil), scripts[index]...)
	}
	return &ScriptedAgent{scripts: copyOfScripts}
}

func (a *ScriptedAgent) RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	a.mu.Lock()
	index := len(a.prompts)
	a.prompts = append(a.prompts, input)
	var events []contracts.AgentStreamEvent
	if index < len(a.scripts) {
		events = append([]contracts.AgentStreamEvent(nil), a.scripts[index]...)
	}
	a.mu.Unlock()
	stream := make(chan contracts.AgentStreamEvent, len(events))
	go func() {
		defer close(stream)
		for _, event := range events {
			if event.Timestamp.IsZero() {
				event.Timestamp = time.Now()
			}
			select {
			case stream <- event:
			case <-ctx.Done():
				return
			}
		}
	}()
	return stream, nil
}

func (a *ScriptedAgent) Prompts() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.prompts...)
}
