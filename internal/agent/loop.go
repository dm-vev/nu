package agent

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/provider"
)

// State is the mutable state for one provider turn.
type State struct {
	Provider   provider.Streamer
	ProviderID string
	API        string
	Model      string
	Emit       func(Event)

	text strings.Builder
}

// TurnInput is one provider turn input.
type TurnInput struct {
	Prompt string
}

func runTurn(ctx context.Context, state *State, input TurnInput) error {
	req := provider.Request{
		Provider: state.ProviderID,
		API:      state.API,
		Model:    state.Model,
		Messages: []provider.Message{{Role: "user", Content: input.Prompt}},
	}
	if err := provider.ValidateRequest(req); err != nil {
		return err
	}

	// Event order mirrors the provider stream contract consumed by JSON/RPC.
	emit(state, Event{Type: "turn_start"})
	ch, err := state.Provider.Stream(ctx, req)
	if err != nil {
		return fmt.Errorf("start provider stream: %w", err)
	}
	for ev := range ch {
		if err := handleProviderEvent(state, ev); err != nil {
			return err
		}
		if ev.Type == provider.EventDone {
			// Final text is emitted once at turn end; deltas already went out live.
			emit(state, Event{Type: "turn_end", Data: map[string]string{"text": state.text.String()}})
			return nil
		}
		if ev.Type == provider.EventError {
			return fmt.Errorf("%w: %s", provider.ErrStream, ev.Message)
		}
	}
	return fmt.Errorf("%w: provider stream closed before done", provider.ErrStream)
}

func handleProviderEvent(state *State, ev provider.Event) error {
	switch ev.Type {
	case provider.EventStart:
		emit(state, Event{Type: "message_start"})
	case provider.EventText:
		state.text.WriteString(ev.Delta)
		emit(state, Event{Type: "message_update", Data: map[string]string{"delta": ev.Delta}})
	case provider.EventDone:
		emit(state, Event{Type: "message_end"})
	case provider.EventError:
	default:
		return fmt.Errorf("%w: unknown event %q", provider.ErrStream, ev.Type)
	}
	return nil
}

func emit(state *State, ev Event) {
	if state.Emit != nil {
		state.Emit(ev)
	}
}
