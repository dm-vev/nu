package agentui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dm-vev/nu/contracts"
)

func TestSDKStreamMapsContentThinkingAndTools(t *testing.T) {
	runner := scriptedRunner{events: []contracts.AgentStreamEvent{
		{Type: contracts.AgentEventThinking, ThinkingStep: "plan"},
		{Type: contracts.AgentEventToolCall, ToolCall: &contracts.ToolCallEvent{ID: "c1", Name: "read", Arguments: `{"path":"a"}`}},
		{Type: contracts.AgentEventToolResult, ToolCall: &contracts.ToolCallEvent{ID: "c1", Name: "read", Result: "ok", Status: "completed"}},
		{Type: contracts.AgentEventContent, Content: "done"},
		{Type: contracts.AgentEventComplete},
	}}
	var events []Event
	controller := New(Options{Runner: runner, Emit: func(event Event) { events = append(events, event) }})
	if err := controller.Prompt(context.Background(), Prompt{Text: "go"}); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"turn_start": false, "message_update": false, "tool_start": false, "tool_end": false, "turn_end": false}
	for _, event := range events {
		if _, ok := want[event.Type]; ok {
			want[event.Type] = true
		}
	}
	for event, found := range want {
		if !found {
			t.Fatalf("events = %#v, missing %s", events, event)
		}
	}
	if got := events[len(events)-1].Data.(map[string]string)["text"]; got != "done" {
		t.Fatalf("final text = %q", got)
	}
}

func TestAbortCancelsSDKRunner(t *testing.T) {
	runner := &blockingRunner{started: make(chan struct{})}
	controller := New(Options{Runner: runner})
	done := make(chan error, 1)
	go func() { done <- controller.Prompt(context.Background(), Prompt{Text: "go"}) }()
	<-runner.started
	controller.Abort()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Prompt error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Prompt did not stop")
	}
}

type scriptedRunner struct{ events []contracts.AgentStreamEvent }

func (r scriptedRunner) RunStream(context.Context, string) (<-chan contracts.AgentStreamEvent, error) {
	stream := make(chan contracts.AgentStreamEvent, len(r.events))
	for _, event := range r.events {
		stream <- event
	}
	close(stream)
	return stream, nil
}

type blockingRunner struct{ started chan struct{} }

func (r *blockingRunner) RunStream(ctx context.Context, _ string) (<-chan contracts.AgentStreamEvent, error) {
	stream := make(chan contracts.AgentStreamEvent)
	close(r.started)
	go func() {
		defer close(stream)
		<-ctx.Done()
	}()
	return stream, nil
}
