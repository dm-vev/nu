package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"

	"google.golang.org/genai"
)

// geminiMessageHistoryBuilder builds Gemini-compatible message history from memory and current prompt
type geminiMessageHistoryBuilder struct {
	logger telemetry.Logger
}

// geminiNewMessageHistoryBuilder creates a new message history builder
func geminiNewMessageHistoryBuilder(logger telemetry.Logger) *geminiMessageHistoryBuilder {
	return &geminiMessageHistoryBuilder{
		logger: logger,
	}
}

// buildContents constructs Gemini contents from memory and current prompt
// Returns contents ready for Gemini API calls, preserving chronological order
func (b *geminiMessageHistoryBuilder) buildContents(ctx context.Context, prompt string, params *contracts.GenerateOptions) []*genai.Content {
	contents := []*genai.Content{}

	// Add memory messages
	if params.Memory != nil {
		memoryMessages, err := params.Memory.GetMessages(ctx)
		if err != nil {
			b.logger.Error(ctx, "Failed to retrieve memory messages", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			// Convert memory messages to Gemini format, preserving chronological order
			for _, msg := range memoryMessages {
				geminiContent := b.convertMemoryMessage(msg)
				if geminiContent != nil {
					contents = append(contents, geminiContent)
				}
			}
		}
	} else {
		// Only append current user message when memory is nil
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: prompt}},
		})
	}

	return contents
}

// convertMemoryMessage converts a memory message to Gemini format
func (b *geminiMessageHistoryBuilder) convertMemoryMessage(msg contracts.Message) *genai.Content {
	switch msg.Role {
	case contracts.RoleUser:
		return &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: msg.Content}},
		}

	case contracts.RoleAssistant:
		if len(msg.ToolCalls) > 0 {
			// Assistant message with tool calls
			var parts []*genai.Part

			// Add text content if present
			if msg.Content != "" {
				parts = append(parts, &genai.Part{Text: msg.Content})
			}

			// Add function calls
			for _, toolCall := range msg.ToolCalls {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
					b.logger.Warn(context.Background(), "Failed to parse tool call arguments", map[string]interface{}{
						"error":     err.Error(),
						"arguments": toolCall.Arguments,
					})
					args = make(map[string]interface{})
				}
				parts = append(parts, &genai.Part{
					ThoughtSignature: toolCall.ThoughtSignature,
					FunctionCall: &genai.FunctionCall{
						Name: toolCall.Name,
						Args: args,
					},
				})
			}

			return &genai.Content{
				Role:  "model",
				Parts: parts,
			}
		} else if msg.Content != "" {
			// Regular assistant message
			return &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: msg.Content}},
			}
		}

	case contracts.MessageRoleTool:
		// Tool messages in Gemini are handled as function responses
		if msg.ToolCallID != "" {
			toolName := "unknown"
			if msg.Metadata != nil {
				if name, ok := msg.Metadata["tool_name"].(string); ok {
					toolName = name
				}
			}
			return &genai.Content{
				Role: "user",
				Parts: []*genai.Part{
					{
						FunctionResponse: &genai.FunctionResponse{
							Name: toolName,
							Response: map[string]any{
								"result": msg.Content,
							},
						},
					},
				},
			}
		}

	case contracts.MessageRoleSystem:
		return &genai.Content{
			Role:  "user", // System instruction is handled separately, other system (like summarized) are passed as user messages
			Parts: []*genai.Part{{Text: fmt.Sprintf("System: %s", msg.Content)}},
		}
	}

	return nil
}
