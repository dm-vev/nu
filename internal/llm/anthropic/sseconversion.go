package anthropic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nu/internal/contracts"
)

// convertAnthropicEventToStreamEvent converts an Anthropic SSE event to our internal StreamEvent
func (c *Client) convertAnthropicEventToStreamEvent(event *SSEEvent, thinkingBlocks map[int]bool, toolBlocks map[int]struct {
	ID        string
	Name      string
	InputJSON strings.Builder
}) (*contracts.StreamEvent, error) {
	if event == nil {
		return nil, nil
	}

	streamEvent := &contracts.StreamEvent{
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	switch event.Type {
	case "message_start":
		var msgStart MessageStartData
		if err := json.Unmarshal(event.Data, &msgStart); err != nil {
			return nil, fmt.Errorf("failed to parse message_start: %w", err)
		}
		streamEvent.Type = contracts.StreamEventMessageStart
		streamEvent.Metadata["message_id"] = msgStart.Message.ID
		streamEvent.Metadata["model"] = msgStart.Message.Model
		streamEvent.Metadata["role"] = msgStart.Message.Role
		streamEvent.Metadata["usage"] = msgStart.Message.Usage

	case "content_block_start":
		var blockStart ContentBlockStartData
		if err := json.Unmarshal(event.Data, &blockStart); err != nil {
			return nil, fmt.Errorf("failed to parse content_block_start: %w", err)
		}
		switch blockStart.ContentBlock.Type {
		case "thinking":
			streamEvent.Type = contracts.StreamEventThinking
			thinkingBlocks[blockStart.Index] = true
			streamEvent.Content = blockStart.ContentBlock.Text
		case "tool_use":
			thinkingBlocks[blockStart.Index] = false
			info := struct {
				ID        string
				Name      string
				InputJSON strings.Builder
			}{ID: blockStart.ContentBlock.ID, Name: blockStart.ContentBlock.Name}
			if len(blockStart.ContentBlock.Input) > 0 {
				argsBytes, _ := json.Marshal(blockStart.ContentBlock.Input)
				info.InputJSON.WriteString(string(argsBytes))
			}
			toolBlocks[blockStart.Index] = info
			return nil, nil
		default:
			streamEvent.Type = contracts.StreamEventContentDelta
			thinkingBlocks[blockStart.Index] = false
			streamEvent.Content = blockStart.ContentBlock.Text
		}
		streamEvent.Metadata["block_index"] = blockStart.Index
		streamEvent.Metadata["block_type"] = blockStart.ContentBlock.Type

	case "content_block_delta":
		var blockDelta ContentBlockDeltaData
		if err := json.Unmarshal(event.Data, &blockDelta); err != nil {
			return nil, fmt.Errorf("failed to parse content_block_delta: %w", err)
		}
		if blockDelta.Delta.Type == "input_json_delta" {
			if info, exists := toolBlocks[blockDelta.Index]; exists {
				info.InputJSON.WriteString(blockDelta.Delta.PartialJSON)
				toolBlocks[blockDelta.Index] = info
			}
			return nil, nil
		}
		if thinkingBlocks[blockDelta.Index] {
			streamEvent.Type = contracts.StreamEventThinking
			streamEvent.Content = blockDelta.Delta.Thinking
		} else {
			streamEvent.Type = contracts.StreamEventContentDelta
			streamEvent.Content = blockDelta.Delta.Text
		}
		streamEvent.Metadata["block_index"] = blockDelta.Index
		streamEvent.Metadata["delta_type"] = blockDelta.Delta.Type

	case "content_block_stop":
		var blockStop ContentBlockStopData
		if err := json.Unmarshal(event.Data, &blockStop); err != nil {
			return nil, fmt.Errorf("failed to parse content_block_stop: %w", err)
		}
		if info, exists := toolBlocks[blockStop.Index]; exists {
			streamEvent.Type = contracts.StreamEventToolUse
			streamEvent.ToolCall = &contracts.ToolCall{
				ID: info.ID, Name: info.Name, Arguments: info.InputJSON.String(),
			}
			streamEvent.Metadata["block_index"] = blockStop.Index
			delete(toolBlocks, blockStop.Index)
		} else {
			streamEvent.Type = contracts.StreamEventContentComplete
			streamEvent.Metadata["block_index"] = blockStop.Index
		}

	case "message_delta":
		var msgDelta MessageDeltaData
		if err := json.Unmarshal(event.Data, &msgDelta); err != nil {
			return nil, fmt.Errorf("failed to parse message_delta: %w", err)
		}
		streamEvent.Type = contracts.StreamEventContentDelta
		streamEvent.Metadata["stop_reason"] = msgDelta.Delta.StopReason
		streamEvent.Metadata["stop_sequence"] = msgDelta.Delta.StopSequence
		streamEvent.Metadata["usage"] = msgDelta.Usage

	case "message_stop":
		streamEvent.Type = contracts.StreamEventMessageStop
	case "ping":
		return nil, nil
	case "error":
		var errorData map[string]interface{}
		if err := json.Unmarshal(event.Data, &errorData); err != nil {
			return nil, fmt.Errorf("failed to parse error event: %w", err)
		}
		streamEvent.Type = contracts.StreamEventError
		streamEvent.Error = fmt.Errorf("anthropic api error: %v", errorData)
		streamEvent.Metadata["error_data"] = errorData
	case "input_json_delta":
		var inputDelta struct {
			Type        string `json:"type"`
			Index       int    `json:"index"`
			PartialJSON string `json:"partial_json"`
		}
		if err := json.Unmarshal(event.Data, &inputDelta); err != nil {
			return nil, fmt.Errorf("failed to parse input_json_delta: %w", err)
		}
		if info, exists := toolBlocks[inputDelta.Index]; exists {
			info.InputJSON.WriteString(inputDelta.PartialJSON)
			toolBlocks[inputDelta.Index] = info
		}
		return nil, nil
	case "done":
		streamEvent.Type = contracts.StreamEventMessageStop
	default:
		streamEvent.Type = contracts.StreamEventContentDelta
		streamEvent.Metadata["unknown_event_type"] = event.Type
		streamEvent.Metadata["raw_data"] = string(event.Data)
	}

	return streamEvent, nil
}
