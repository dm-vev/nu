package anthropic

import (
	"context"
	"testing"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

func TestAnthropicMessageHistoryBuilder_BuildMessages(t *testing.T) {
	logger := telemetry.NewLogger()
	builder := anthropicNewMessageHistoryBuilder(logger)

	tests := []struct {
		name     string
		prompt   string
		params   *contracts.GenerateOptions
		expected int
	}{
		{
			name:     "no memory",
			prompt:   "Hello",
			params:   &contracts.GenerateOptions{},
			expected: 1,
		},
		{
			name:   "with system message",
			prompt: "Hello",
			params: &contracts.GenerateOptions{
				SystemMessage: "You are helpful",
			},
			expected: 1,
		},
		{
			name:   "with memory",
			prompt: "Continue",
			params: &contracts.GenerateOptions{
				Memory: &anthropicMockMemory{
					messages: []contracts.Message{
						{Role: contracts.RoleUser, Content: "Hi"},
						{Role: contracts.RoleAssistant, Content: "Hello!"},
						{Role: contracts.RoleUser, Content: "Continue"}, // Agent adds current prompt to memory by default
					},
				},
			},
			expected: 3, // 2 from memory + 1 current prompt
		},
		{
			name:   "with memory including system",
			prompt: "Continue",
			params: &contracts.GenerateOptions{
				Memory: &anthropicMockMemory{
					messages: []contracts.Message{
						{Role: contracts.MessageRoleSystem, Content: "Old system"},
						{Role: contracts.RoleUser, Content: "Hi"},
						{Role: contracts.RoleAssistant, Content: "Hello!"},
						{Role: contracts.RoleUser, Content: "Continue"}, // Agent adds current prompt to memory by default
					},
				},
			},
			expected: 4, // 3 from memory + 1 current prompt
		},
		{
			name:   "with tool calls and results",
			prompt: "What's next?",
			params: &contracts.GenerateOptions{
				Memory: &anthropicMockMemory{
					messages: []contracts.Message{
						{Role: contracts.RoleUser, Content: "Get weather"},
						{Role: contracts.RoleAssistant, Content: "I'll check the weather", ToolCalls: []contracts.ToolCall{
							{ID: "call_123", Name: "get_weather", Arguments: `{"location": "NYC"}`},
						}},
						{Role: contracts.MessageRoleTool, Content: "Sunny, 72°F", ToolCallID: "call_123"},
						{Role: contracts.RoleUser, Content: "What's next?"}, // Agent adds current prompt to memory by default
					},
				},
			},
			expected: 4, // 3 from memory + 1 current prompt
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := builder.buildMessages(context.Background(), tt.prompt, tt.params)
			if len(messages) != tt.expected {
				t.Errorf("Expected %d messages, got %d", tt.expected, len(messages))
			}
		})
	}
}

// anthropicMockMemory is a simple mock implementation for testing
type anthropicMockMemory struct {
	messages []contracts.Message
}

func (m *anthropicMockMemory) AddMessage(ctx context.Context, message contracts.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *anthropicMockMemory) GetMessages(ctx context.Context, options ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	return m.messages, nil
}

func (m *anthropicMockMemory) Clear(ctx context.Context) error {
	m.messages = []contracts.Message{}
	return nil
}
