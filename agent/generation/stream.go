package generation

import (
	"context"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
)

func sendEvent(ctx context.Context, eventChan chan<- contracts.AgentStreamEvent, event contracts.AgentStreamEvent) bool {
	select {
	case eventChan <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// Stream forwards provider events and records the completed exchange in memory.
func (s *Service) Stream(ctx context.Context, input string, allTools []contracts.Tool, streamingLLM contracts.StreamingLLM, eventChan chan<- contracts.AgentStreamEvent) (int64, error) {
	options := s.streamOptions()
	ctx = withStreamForwarder(ctx, eventChan)
	llmEventChan, err := s.startStream(ctx, input, allTools, streamingLLM, options)
	if err != nil {
		return 0, err
	}

	var content strings.Builder
	var toolCalls []contracts.ToolCall
	toolResults := make(map[string]string)
	var finalError error

	for llmEvent := range llmEventChan {
		if llmEvent.Type == contracts.StreamEventContentDelta {
			content.WriteString(llmEvent.Content)
		}
		if llmEvent.Type == contracts.StreamEventToolUse && llmEvent.ToolCall != nil {
			toolCalls = append(toolCalls, *llmEvent.ToolCall)
		}
		if llmEvent.Type == contracts.StreamEventToolResult && llmEvent.ToolCall != nil {
			toolResults[llmEvent.ToolCall.ID] = llmEvent.Content
		}
		if llmEvent.Error != nil {
			finalError = llmEvent.Error
		}
		if !sendEvent(ctx, eventChan, ConvertLLMEventToAgentEvent(llmEvent, allTools)) {
			return int64(content.Len()), finalError
		}
	}

	saveStreamMessages(ctx, s.Memory, s.Logger, content.String(), toolCalls, toolResults)
	sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
		Type:      contracts.AgentEventComplete,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"total_content_length": content.Len(),
			"had_error":            finalError != nil,
		},
	})
	return int64(content.Len()), finalError
}
