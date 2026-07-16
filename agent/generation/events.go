package generation

import "github.com/dm-vev/nu/contracts"

// getToolMetadata retrieves display name and internal flag for a tool
func getToolMetadata(toolName string, tools []contracts.Tool) (displayName string, internal bool) {
	displayName = toolName
	internal = false

	for _, tool := range tools {
		if tool.Name() == toolName {
			if toolWithDisplayName, ok := tool.(contracts.ToolWithDisplayName); ok {
				if dn := toolWithDisplayName.DisplayName(); dn != "" {
					displayName = dn
				}
			}
			if internalTool, ok := tool.(contracts.InternalTool); ok {
				internal = internalTool.Internal()
			}
			break
		}
	}

	return displayName, internal
}

// convertLLMEventToAgentEvent converts LLM events to agent events
func ConvertLLMEventToAgentEvent(llmEvent contracts.StreamEvent, tools []contracts.Tool) contracts.AgentStreamEvent {
	agentEvent := contracts.AgentStreamEvent{
		Timestamp: llmEvent.Timestamp,
		Metadata:  llmEvent.Metadata,
	}

	// Convert event types
	switch llmEvent.Type {
	case contracts.StreamEventMessageStart:
		agentEvent.Type = contracts.AgentEventContent
		agentEvent.Content = llmEvent.Content

	case contracts.StreamEventContentDelta:
		agentEvent.Type = contracts.AgentEventContent
		agentEvent.Content = llmEvent.Content

	case contracts.StreamEventContentComplete:
		agentEvent.Type = contracts.AgentEventContent
		agentEvent.Content = llmEvent.Content

	case contracts.StreamEventThinking:
		agentEvent.Type = contracts.AgentEventThinking
		agentEvent.ThinkingStep = llmEvent.Content

	case contracts.StreamEventToolUse:
		agentEvent.Type = contracts.AgentEventToolCall
		if llmEvent.ToolCall != nil {
			displayName, internal := getToolMetadata(llmEvent.ToolCall.Name, tools)
			agentEvent.ToolCall = &contracts.ToolCallEvent{
				ID:          llmEvent.ToolCall.ID,
				Name:        llmEvent.ToolCall.Name,
				DisplayName: displayName,
				Internal:    internal,
				Arguments:   llmEvent.ToolCall.Arguments,
				Status:      "received",
			}
		}

	case contracts.StreamEventToolResult:
		agentEvent.Type = contracts.AgentEventToolResult
		if llmEvent.ToolCall != nil {
			displayName, internal := getToolMetadata(llmEvent.ToolCall.Name, tools)
			agentEvent.ToolCall = &contracts.ToolCallEvent{
				ID:          llmEvent.ToolCall.ID,
				Name:        llmEvent.ToolCall.Name,
				DisplayName: displayName,
				Internal:    internal,
				Arguments:   llmEvent.ToolCall.Arguments,
				Result:      llmEvent.Content, // Tool result is in Content field of StreamEvent
				Status:      "completed",
			}
		}

	case contracts.StreamEventError:
		agentEvent.Type = contracts.AgentEventError
		agentEvent.Error = llmEvent.Error

	case contracts.StreamEventMessageStop:
		agentEvent.Type = contracts.AgentEventContent
		agentEvent.Content = llmEvent.Content

	default:
		// Unknown event type, treat as content
		agentEvent.Type = contracts.AgentEventContent
		agentEvent.Content = llmEvent.Content
	}

	return agentEvent
}
