package agentui

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
)

func consumeStream(ctx context.Context, runner contracts.StreamingAgent, input string, emit func(Event)) error {
	stream, err := runner.RunStream(ctx, input)
	if err != nil {
		return fmt.Errorf("start SDK agent stream: %w", err)
	}
	emitEvent(emit, Event{Type: "turn_start"})
	emitEvent(emit, Event{Type: "message_start"})
	var text strings.Builder
	var streamErr error
	for event := range stream {
		switch event.Type {
		case contracts.AgentEventContent:
			if event.Content == "" || boundaryContent(event.Metadata) {
				continue
			}
			text.WriteString(event.Content)
			emitEvent(emit, Event{Type: "message_update", Data: map[string]string{"delta": event.Content}})
		case contracts.AgentEventThinking:
			if event.ThinkingStep != "" {
				emitEvent(emit, Event{Type: "message_update", Data: map[string]string{"kind": "thinking", "delta": event.ThinkingStep, "thinking_delta": event.ThinkingStep}})
			}
		case contracts.AgentEventToolCall:
			emitToolStart(emit, event.ToolCall)
		case contracts.AgentEventToolResult:
			emitToolEnd(emit, event.ToolCall)
		case contracts.AgentEventError:
			if event.Error != nil && streamErr == nil {
				streamErr = event.Error
			}
		case contracts.AgentEventComplete:
		}
	}
	if streamErr != nil {
		return streamErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	emitEvent(emit, Event{Type: "message_end"})
	emitEvent(emit, Event{Type: "turn_end", Data: map[string]string{"text": text.String()}})
	return nil
}

func emitToolStart(emit func(Event), call *contracts.ToolCallEvent) {
	if call == nil {
		return
	}
	data := map[string]string{"id": call.ID, "name": call.Name, "arguments": call.Arguments}
	emitEvent(emit, Event{Type: "tool_call_start", Data: data})
	emitEvent(emit, Event{Type: "tool_call_end", Data: data})
	emitEvent(emit, Event{Type: "tool_start", Data: data})
}

func emitToolEnd(emit func(Event), call *contracts.ToolCallEvent) {
	if call == nil {
		return
	}
	isError := "false"
	if call.Status == "error" {
		isError = "true"
	}
	emitEvent(emit, Event{Type: "tool_end", Data: map[string]string{"id": call.ID, "name": call.Name, "result": call.Result, "error": isError}})
}

func boundaryContent(metadata map[string]any) bool {
	if metadata == nil {
		return false
	}
	_, beforeTools := metadata["before_tools"]
	_, iterationBoundary := metadata["iteration_boundary"]
	return beforeTools || iterationBoundary
}

func emitEvent(emit func(Event), event Event) {
	if emit != nil {
		emit(event)
	}
}
