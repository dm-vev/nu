package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"nu/internal/provider"
	"nu/internal/testkit"
)

func TestNUF050TextOnlyTurnEnds(t *testing.T) {
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "hello"},
		provider.Event{Type: provider.EventDone, StopReason: "stop"},
	)
	var events []Event
	a := New(Options{
		Provider: fake,
		Emit: func(ev Event) {
			events = append(events, ev)
		},
	})

	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}
	requests := fake.Requests()
	if len(requests) != 1 {
		t.Fatalf("Provider requests = %d, want 1", len(requests))
	}
	if requests[0].Messages[0].Content != "hi" {
		t.Fatalf("Provider prompt = %q, want hi", requests[0].Messages[0].Content)
	}
	if got := events[len(events)-1].Type; got != "turn_end" {
		t.Fatalf("Last event = %q, want turn_end", got)
	}
}

func TestNUF050ToolCallFeedsResultBackToProvider(t *testing.T) {
	fake := testkit.NewScriptedProviderScripts(
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventToolCallStart, Index: 0, ToolCallID: "call-1", ToolName: "fake"},
			{Type: provider.EventToolCallDelta, Index: 0, Delta: `{"input":"hi"}`},
			{Type: provider.EventToolCallEnd, Index: 0},
			{Type: provider.EventDone, StopReason: "tool_use"},
		},
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventText, Delta: "done"},
			{Type: provider.EventDone, StopReason: "stop"},
		},
	)
	var events []Event
	a := New(Options{
		Provider: fake,
		Tools: map[string]ToolFunc{
			"fake": func(_ context.Context, call ToolCall) (ToolResult, error) {
				if call.ID != "call-1" {
					t.Fatalf("Tool call id = %q, want call-1", call.ID)
				}
				if call.Arguments != `{"input":"hi"}` {
					t.Fatalf("Tool arguments = %q, want raw JSON", call.Arguments)
				}
				return ToolResult{Content: "tool result"}, nil
			},
		},
		Emit: func(ev Event) {
			events = append(events, ev)
		},
	})

	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}
	requests := fake.Requests()
	if len(requests) != 2 {
		t.Fatalf("Provider requests = %d, want 2", len(requests))
	}
	lastMessage := requests[1].Messages[len(requests[1].Messages)-1]
	if lastMessage.Role != "tool" || lastMessage.ToolCallID != "call-1" || lastMessage.Content != "tool result" {
		t.Fatalf("Second request last message = %#v, want tool result", lastMessage)
	}
	if got := events[len(events)-1].Data.(map[string]string)["text"]; got != "done" {
		t.Fatalf("Final text = %q, want done", got)
	}
}

func TestNUF050AbortStopsProviderAndTools(t *testing.T) {
	fake := &blockingProvider{started: make(chan struct{})}
	a := New(Options{Provider: fake})

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Prompt(context.Background(), Prompt{Text: "hi"})
	}()
	<-fake.started

	a.Abort()
	err := <-errCh
	if !errors.Is(err, provider.ErrStream) {
		t.Fatalf("Prompt error = %v, want provider ErrStream", err)
	}
}

func TestNUF050MissingToolFails(t *testing.T) {
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventToolCallStart, Index: 0, ToolCallID: "call-1", ToolName: "missing"},
		provider.Event{Type: provider.EventToolCallEnd, Index: 0},
		provider.Event{Type: provider.EventDone, StopReason: "tool_use"},
	)
	a := New(Options{Provider: fake})

	err := a.Prompt(context.Background(), Prompt{Text: "hi"})
	if err == nil || !strings.Contains(err.Error(), "missing tool") {
		t.Fatalf("Prompt error = %v, want missing tool", err)
	}
}

type blockingProvider struct {
	started chan struct{}
}

func (p *blockingProvider) Stream(ctx context.Context, _ provider.Request) (<-chan provider.Event, error) {
	close(p.started)
	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}
