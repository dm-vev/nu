package ollama

import (
	"context"
	"encoding/json"
	"fmt"

	"nu/internal/llm"
)

// Chat performs a chat completion with messages
func (c *Client) Chat(ctx context.Context, messages []llm.Message, params *llm.GenerateParams) (string, error) {
	// Convert messages to Ollama format
	var chatMessages []ChatMessage
	for _, msg := range messages {
		chatMessages = append(chatMessages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Create request
	req := ChatRequest{
		Model:    c.Model,
		Messages: chatMessages,
		Stream:   false,
		Options: &Options{
			Temperature: params.Temperature,
			TopP:        params.TopP,
			Stop:        params.StopSequences,
		},
	}

	// Make request
	resp, err := c.makeRequest(ctx, "/api/chat", req)
	if err != nil {
		return "", fmt.Errorf("failed to chat: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(resp, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal chat response: %w", err)
	}

	return chatResp.Message.Content, nil
}
