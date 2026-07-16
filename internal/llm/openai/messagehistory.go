package openai

import (
	"context"

	"nu/internal/contracts"
	"nu/internal/telemetry"

	"github.com/openai/openai-go/v2"
)

// openAIMessageHistoryBuilder builds OpenAI-compatible message history from memory and current prompt
type openAIMessageHistoryBuilder struct {
	logger telemetry.Logger
}

// openAINewMessageHistoryBuilder creates a new message history builder
func openAINewMessageHistoryBuilder(logger telemetry.Logger) *openAIMessageHistoryBuilder {
	return &openAIMessageHistoryBuilder{
		logger: logger,
	}
}

// buildMessages constructs OpenAI messages from memory and current prompt
// Returns messages ready for OpenAI API calls, preserving chronological order
func (b *openAIMessageHistoryBuilder) buildMessages(ctx context.Context, prompt string, memory contracts.Memory) []openai.ChatCompletionMessageParamUnion {
	messages := []openai.ChatCompletionMessageParamUnion{}

	// Add memory messages
	if memory != nil {
		memoryMessages, err := memory.GetMessages(ctx)
		if err != nil {
			b.logger.Error(ctx, "Failed to retrieve memory messages", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			// Convert memory messages to OpenAI format, preserving chronological order
			for _, msg := range memoryMessages {
				openaiMsg := b.convertMemoryMessage(msg)
				if openaiMsg != nil {
					messages = append(messages, *openaiMsg)
				}
			}
		}
	} else {
		// Only append current user message when memory is nil
		messages = append(messages, openai.UserMessage(prompt))
	}

	return messages
}

// convertMemoryMessage converts a memory message to OpenAI format
func (b *openAIMessageHistoryBuilder) convertMemoryMessage(msg contracts.Message) *openai.ChatCompletionMessageParamUnion {
	switch msg.Role {
	case contracts.RoleUser:
		userMsg := openai.UserMessage(msg.Content)
		return &userMsg

	case contracts.RoleAssistant:
		if len(msg.ToolCalls) > 0 {
			// Assistant message with tool calls
			var toolCalls []openai.ChatCompletionMessageToolCallUnion

			for _, toolCall := range msg.ToolCalls {
				toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnion{
					ID:   toolCall.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
					},
				})
			}

			// Create assistant message with tool calls
			assistantMsg := openai.ChatCompletionMessage{
				Role:      "assistant",
				Content:   msg.Content,
				ToolCalls: toolCalls,
			}
			param := assistantMsg.ToParam()
			return &param
		} else if msg.Content != "" {
			// Regular assistant message
			assistantMsg := openai.AssistantMessage(msg.Content)
			return &assistantMsg
		}

	case contracts.MessageRoleTool:
		if msg.ToolCallID != "" {
			toolMsg := openai.ToolMessage(msg.Content, msg.ToolCallID)
			return &toolMsg
		}

	case contracts.MessageRoleSystem:
		// Convert system messages from memory to OpenAI system messages
		systemMsg := openai.SystemMessage(msg.Content)
		return &systemMsg
	}

	return nil
}
