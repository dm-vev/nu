package agent

import (
	"context"
	"errors"
	"fmt"
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

func TestNUF050ProviderRequestIncludesToolDefinitions(t *testing.T) {
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone, StopReason: "stop"},
	)
	a := New(Options{
		Provider: fake,
		ToolDefs: []provider.ToolDefinition{{
			Name:        "bash",
			Description: "run shell",
			Parameters:  map[string]any{"type": "object"},
		}},
	})

	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}
	requests := fake.Requests()
	if len(requests) != 1 || len(requests[0].Tools) != 1 || requests[0].Tools[0].Name != "bash" {
		t.Fatalf("requests = %#v, want bash tool definition", requests)
	}
}

func TestAgentPromptIncludesPreviousTurns(t *testing.T) {
	fake := testkit.NewScriptedProviderScripts(
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventText, Delta: "first answer"},
			{Type: provider.EventDone, StopReason: "stop"},
		},
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventText, Delta: "second answer"},
			{Type: provider.EventDone, StopReason: "stop"},
		},
	)
	a := New(Options{Provider: fake})

	if err := a.Prompt(context.Background(), Prompt{Text: "what did I ask?"}); err != nil {
		t.Fatalf("first Prompt error = %v", err)
	}
	if err := a.Prompt(context.Background(), Prompt{Text: "what did you answer?"}); err != nil {
		t.Fatalf("second Prompt error = %v", err)
	}

	messages := fake.Requests()[1].Messages
	want := []provider.Message{
		{Role: "user", Content: "what did I ask?"},
		{Role: "assistant", Content: "first answer"},
		{Role: "user", Content: "what did you answer?"},
	}
	if fmt.Sprint(messages) != fmt.Sprint(want) {
		t.Fatalf("second request messages = %#v, want %#v", messages, want)
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
	messages := requests[1].Messages
	assistantMessage := messages[len(messages)-2]
	if assistantMessage.Role != "assistant" ||
		assistantMessage.ToolCallID != "call-1" ||
		assistantMessage.Name != "fake" ||
		assistantMessage.Content != `{"input":"hi"}` {
		t.Fatalf("Second request assistant tool message = %#v, want tool call context", assistantMessage)
	}
	lastMessage := requests[1].Messages[len(requests[1].Messages)-1]
	if lastMessage.Role != "tool" || lastMessage.ToolCallID != "call-1" || lastMessage.Content != "tool result" {
		t.Fatalf("Second request last message = %#v, want tool result", lastMessage)
	}
	if got := events[len(events)-1].Data.(map[string]string)["text"]; got != "done" {
		t.Fatalf("Final text = %q, want done", got)
	}
	var toolStart map[string]string
	var toolEnd map[string]string
	for _, ev := range events {
		switch ev.Type {
		case "tool_start":
			toolStart = ev.Data.(map[string]string)
		case "tool_end":
			toolEnd = ev.Data.(map[string]string)
		}
	}
	if toolStart["arguments"] != `{"input":"hi"}` {
		t.Fatalf("tool_start data = %#v, want arguments", toolStart)
	}
	if toolEnd["result"] != "tool result" || toolEnd["error"] != "false" {
		t.Fatalf("tool_end data = %#v, want result and success flag", toolEnd)
	}
}

func TestNUF050ThinkingDeltaEmitsStructuredMessageUpdate(t *testing.T) {
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventThinking, Delta: "checking state"},
		provider.Event{Type: provider.EventText, Delta: "answer"},
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
	foundThinking := false
	for _, ev := range events {
		if ev.Type != "message_update" {
			continue
		}
		data := ev.Data.(map[string]string)
		if data["kind"] == "thinking" && data["thinking_delta"] == "checking state" {
			foundThinking = true
		}
	}
	if !foundThinking {
		t.Fatalf("events = %#v, want thinking message_update", events)
	}
	if got := events[len(events)-1].Data.(map[string]string)["text"]; got != "answer" {
		t.Fatalf("Final text = %q, want answer without thinking", got)
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

func TestNUF050RejectsMalformedToolCallStream(t *testing.T) {
	tests := []struct {
		name   string
		events []provider.Event
		want   string
	}{
		{
			name: "missing id",
			events: []provider.Event{
				{Type: provider.EventStart},
				{Type: provider.EventToolCallStart, Index: 0, ToolName: "fake"},
			},
			want: "missing tool call id",
		},
		{
			name: "delta after end",
			events: []provider.Event{
				{Type: provider.EventStart},
				{Type: provider.EventToolCallStart, Index: 0, ToolCallID: "call-1", ToolName: "fake"},
				{Type: provider.EventToolCallEnd, Index: 0},
				{Type: provider.EventToolCallDelta, Index: 0, Delta: "{}"},
			},
			want: "tool call delta after end",
		},
		{
			name: "duplicate end",
			events: []provider.Event{
				{Type: provider.EventStart},
				{Type: provider.EventToolCallStart, Index: 0, ToolCallID: "call-1", ToolName: "fake"},
				{Type: provider.EventToolCallEnd, Index: 0},
				{Type: provider.EventToolCallEnd, Index: 0},
			},
			want: "duplicate tool call end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := testkit.NewScriptedProvider(tt.events...)
			a := New(Options{Provider: fake})

			err := a.Prompt(context.Background(), Prompt{Text: "hi"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Prompt error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestAgentSetModelAffectsNextPrompt(t *testing.T) {
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone, StopReason: "stop"},
	)
	a := New(Options{Provider: fake})

	if err := a.SetModel("openai", "chat", "gpt-test"); err != nil {
		t.Fatalf("SetModel error = %v", err)
	}
	cfg := a.Config()
	if cfg.ProviderID != "openai" || cfg.API != "chat" || cfg.Model != "gpt-test" {
		t.Fatalf("Config = %#v, want updated labels", cfg)
	}
	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}

	requests := fake.Requests()
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	if requests[0].Provider != "openai" || requests[0].API != "chat" || requests[0].Model != "gpt-test" {
		t.Fatalf("provider request = %#v, want updated labels", requests[0])
	}
}

func TestAgentSetProviderModelSwitchesStreamer(t *testing.T) {
	first := testkit.NewScriptedProvider()
	second := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone, StopReason: "stop"},
	)
	a := New(Options{Provider: first})

	if err := a.SetProviderModel(second, "openai", "chat", "gpt-test"); err != nil {
		t.Fatalf("SetProviderModel error = %v", err)
	}
	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}

	if got := len(first.Requests()); got != 0 {
		t.Fatalf("first provider requests = %d, want 0", got)
	}
	requests := second.Requests()
	if len(requests) != 1 {
		t.Fatalf("second provider requests = %d, want 1", len(requests))
	}
	if requests[0].Provider != "openai" || requests[0].Model != "gpt-test" {
		t.Fatalf("second provider request = %#v, want switched model", requests[0])
	}
}

func TestAgentRetriesRateLimitBeforeFailing(t *testing.T) {
	fake := testkit.NewScriptedProviderErrors(
		[]error{
			fmt.Errorf("%w: wait", provider.ErrRateLimit),
			fmt.Errorf("%w: wait", provider.ErrRateLimit),
			nil,
		},
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventText, Delta: "ok"},
			{Type: provider.EventDone, StopReason: "stop"},
		},
	)
	var rateLimitEvents int
	a := New(Options{
		Provider: fake,
		Emit: func(ev Event) {
			if ev.Type == "rate_limit" {
				rateLimitEvents++
			}
		},
	})

	if err := a.Prompt(context.Background(), Prompt{Text: "hi"}); err != nil {
		t.Fatalf("Prompt error = %v", err)
	}
	if got := len(fake.Requests()); got != 3 {
		t.Fatalf("requests = %d, want 3", got)
	}
	if rateLimitEvents != 2 {
		t.Fatalf("rate limit events = %d, want 2", rateLimitEvents)
	}
}

func TestAgentStopsAfterFiveRateLimitRetries(t *testing.T) {
	failures := []error{
		fmt.Errorf("%w: wait", provider.ErrRateLimit),
		fmt.Errorf("%w: wait", provider.ErrRateLimit),
		fmt.Errorf("%w: wait", provider.ErrRateLimit),
		fmt.Errorf("%w: wait", provider.ErrRateLimit),
		fmt.Errorf("%w: wait", provider.ErrRateLimit),
		fmt.Errorf("%w: wait", provider.ErrRateLimit),
	}
	fake := testkit.NewScriptedProviderErrors(failures)
	var rateLimitEvents int
	a := New(Options{
		Provider: fake,
		Emit: func(ev Event) {
			if ev.Type == "rate_limit" {
				rateLimitEvents++
			}
		},
	})

	err := a.Prompt(context.Background(), Prompt{Text: "hi"})
	if !errors.Is(err, provider.ErrRateLimit) {
		t.Fatalf("Prompt error = %v, want rate limit", err)
	}
	if got := len(fake.Requests()); got != 6 {
		t.Fatalf("requests = %d, want initial request plus 5 retries", got)
	}
	if rateLimitEvents != 5 {
		t.Fatalf("rate limit events = %d, want 5", rateLimitEvents)
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
