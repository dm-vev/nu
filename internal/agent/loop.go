package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"nu/internal/provider"
)

// State is the mutable state for one provider turn.
type State struct {
	Provider   provider.Streamer
	ProviderID string
	API        string
	Model      string
	Tools      map[string]ToolFunc
	Emit       func(Event)

	messages  []provider.Message
	toolCalls map[int]*pendingToolCall
	text      strings.Builder
}

// TurnInput is one provider turn input.
type TurnInput struct {
	Prompt string
}

type pendingToolCall struct {
	call      ToolCall
	arguments strings.Builder
	done      bool
}

func runTurn(ctx context.Context, state *State, input TurnInput) error {
	state.messages = []provider.Message{{Role: "user", Content: input.Prompt}}

	// Event order mirrors the provider stream contract consumed by JSON/RPC.
	emit(state, Event{Type: "turn_start"})
	for {
		stopReason, err := runProviderStream(ctx, state)
		if err != nil {
			return err
		}
		if stopReason != "tool_use" {
			// Final text is emitted once at turn end; deltas already went out live.
			emit(state, Event{Type: "turn_end", Data: map[string]string{"text": state.text.String()}})
			return nil
		}

		// A tool-use stop is not terminal; tool results become the next provider input.
		results, err := executeToolCalls(ctx, state)
		if err != nil {
			return err
		}
		state.messages = append(state.messages, results...)
	}
}

func runProviderStream(ctx context.Context, state *State) (string, error) {
	req := provider.Request{
		Provider: state.ProviderID,
		API:      state.API,
		Model:    state.Model,
		Messages: append([]provider.Message(nil), state.messages...),
	}
	if err := provider.ValidateRequest(req); err != nil {
		return "", err
	}

	// Tool-call buffers are scoped to one provider response.
	state.toolCalls = make(map[int]*pendingToolCall)
	ch, err := state.Provider.Stream(ctx, req)
	if err != nil {
		return "", fmt.Errorf("start provider stream: %w", err)
	}
	for ev := range ch {
		if err := handleProviderEvent(state, ev); err != nil {
			return "", err
		}
		if ev.Type == provider.EventDone {
			return ev.StopReason, nil
		}
		if ev.Type == provider.EventError {
			return "", fmt.Errorf("%w: %s", provider.ErrStream, ev.Message)
		}
	}
	return "", fmt.Errorf("%w: provider stream closed before done", provider.ErrStream)
}

func handleProviderEvent(state *State, ev provider.Event) error {
	switch ev.Type {
	case provider.EventStart:
		emit(state, Event{Type: "message_start"})
	case provider.EventText:
		state.text.WriteString(ev.Delta)
		emit(state, Event{Type: "message_update", Data: map[string]string{"delta": ev.Delta}})
	case provider.EventToolCallStart:
		if _, exists := state.toolCalls[ev.Index]; exists {
			return fmt.Errorf("%w: duplicate tool call index %d", provider.ErrStream, ev.Index)
		}
		state.toolCalls[ev.Index] = &pendingToolCall{
			call: ToolCall{ID: ev.ToolCallID, Name: ev.ToolName},
		}
		emit(state, Event{Type: "tool_call_start", Data: map[string]string{"id": ev.ToolCallID, "name": ev.ToolName}})
	case provider.EventToolCallDelta:
		call, ok := state.toolCalls[ev.Index]
		if !ok {
			return fmt.Errorf("%w: tool call delta before start %d", provider.ErrStream, ev.Index)
		}
		call.arguments.WriteString(ev.Delta)
	case provider.EventToolCallEnd:
		call, ok := state.toolCalls[ev.Index]
		if !ok {
			return fmt.Errorf("%w: tool call end before start %d", provider.ErrStream, ev.Index)
		}
		call.call.Arguments = call.arguments.String()
		call.done = true
		emit(state, Event{Type: "tool_call_end", Data: map[string]string{"id": call.call.ID, "name": call.call.Name}})
	case provider.EventDone:
		emit(state, Event{Type: "message_end"})
	case provider.EventError:
	default:
		return fmt.Errorf("%w: unknown event %q", provider.ErrStream, ev.Type)
	}
	return nil
}

func executeToolCalls(ctx context.Context, state *State) ([]provider.Message, error) {
	if len(state.toolCalls) == 0 {
		return nil, fmt.Errorf("%w: tool_use without tool calls", provider.ErrStream)
	}

	indexes := make([]int, 0, len(state.toolCalls))
	for index := range state.toolCalls {
		indexes = append(indexes, index)
	}
	// Provider chunks may interleave; result messages still follow tool-call index order.
	sort.Ints(indexes)

	results := make([]provider.Message, 0, len(indexes))
	for _, index := range indexes {
		pending := state.toolCalls[index]
		if !pending.done {
			return nil, fmt.Errorf("%w: unfinished tool call %s", provider.ErrStream, pending.call.ID)
		}
		tool := state.Tools[pending.call.Name]
		if tool == nil {
			return nil, fmt.Errorf("missing tool %q", pending.call.Name)
		}
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("execute tool %s: %w", pending.call.Name, err)
		}

		emit(state, Event{Type: "tool_start", Data: map[string]string{"id": pending.call.ID, "name": pending.call.Name}})
		result, err := tool(ctx, pending.call)
		if err != nil {
			return nil, fmt.Errorf("execute tool %s: %w", pending.call.Name, err)
		}
		emit(state, Event{Type: "tool_end", Data: map[string]string{"id": pending.call.ID, "name": pending.call.Name}})
		results = append(results, provider.Message{
			Role:       "tool",
			Content:    result.Content,
			ToolCallID: pending.call.ID,
			Name:       pending.call.Name,
		})
	}
	return results, nil
}

func emit(state *State, ev Event) {
	if state.Emit != nil {
		state.Emit(ev)
	}
}
