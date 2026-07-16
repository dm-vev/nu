package azureopenai

import (
	"context"
	"testing"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

func TestAzureOpenAIMessageHistoryBuilder_BuildMessages(t *testing.T) {
	logger := telemetry.NewLogger()
	builder := azureOpenAINewMessageHistoryBuilder(logger)

	tests := []struct {
		name     string
		prompt   string
		memory   contracts.Memory
		expected int
	}{
		{
			name:     "no memory",
			prompt:   "Hello",
			memory:   nil,
			expected: 1,
		},
		{
			name:   "with memory",
			prompt: "Continue",
			memory: &azureOpenAIMockMemory{
				messages: []contracts.Message{
					{Role: contracts.RoleUser, Content: "Hi"},
					{Role: contracts.RoleAssistant, Content: "Hello!"},
				},
			},
			expected: 2,
		},
		{
			name:   "with memory including system",
			prompt: "Continue",
			memory: &azureOpenAIMockMemory{
				messages: []contracts.Message{
					{Role: contracts.MessageRoleSystem, Content: "Old system"},
					{Role: contracts.RoleUser, Content: "Hi"},
					{Role: contracts.RoleAssistant, Content: "Hello!"},
				},
			},
			expected: 3,
		},
		{
			name:   "with tool calls and results",
			prompt: "What's next?",
			memory: &azureOpenAIMockMemory{
				messages: []contracts.Message{
					{Role: contracts.RoleUser, Content: "Get weather"},
					{Role: contracts.RoleAssistant, Content: "I'll check the weather", ToolCalls: []contracts.ToolCall{
						{ID: "call_123", Name: "get_weather", Arguments: `{"location": "NYC"}`},
					}},
					{Role: contracts.MessageRoleTool, Content: "Sunny, 72°F", ToolCallID: "call_123"},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := builder.buildMessages(context.Background(), tt.prompt, tt.memory)
			if len(messages) != tt.expected {
				t.Errorf("Expected %d messages, got %d", tt.expected, len(messages))
			}
		})
	}
}

// azureOpenAIMockMemory is a simple mock implementation for testing
type azureOpenAIMockMemory struct {
	messages []contracts.Message
}

func (m *azureOpenAIMockMemory) AddMessage(ctx context.Context, message contracts.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *azureOpenAIMockMemory) GetMessages(ctx context.Context, options ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	return m.messages, nil
}

func (m *azureOpenAIMockMemory) Clear(ctx context.Context) error {
	m.messages = []contracts.Message{}
	return nil
}
