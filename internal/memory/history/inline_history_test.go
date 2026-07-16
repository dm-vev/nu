package history

import (
	"context"
	"testing"

	"nu/internal/contracts"
)

func TestBuildInlineHistoryPrompt(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		history  []contracts.Message
		expected string
	}{
		{
			name:     "no memory",
			prompt:   "Hello",
			history:  nil,
			expected: "User: Hello",
		},
		{
			name:     "empty memory",
			prompt:   "Hello",
			history:  []contracts.Message{},
			expected: "Hello",
		},
		{
			name:   "system message first",
			prompt: "Hello",
			history: []contracts.Message{
				{Role: contracts.RoleUser, Content: "Previous question"},
				{Role: contracts.MessageRoleSystem, Content: "You are a helpful assistant"},
				{Role: contracts.RoleAssistant, Content: "Previous answer"},
			},
			expected: "System: You are a helpful assistant\nUser: Previous question\nAssistant: Previous answer\nUser: Hello",
		},
		{
			name:   "conversation with tools",
			prompt: "What's the status?",
			history: []contracts.Message{
				{Role: contracts.RoleUser, Content: "Check the database"},
				{Role: contracts.RoleAssistant, Content: "I'll check the database for you"},
				{
					Role:       contracts.MessageRoleTool,
					Content:    "Database is running",
					ToolCallID: "tool_123",
					Metadata:   map[string]interface{}{"tool_name": "db_check"},
				},
				{Role: contracts.RoleAssistant, Content: "The database is running normally"},
			},
			expected: "User: Check the database\nAssistant: I'll check the database for you\nTool db_check result: Database is running\nAssistant: The database is running normally\nUser: What's the status?",
		},
		{
			name:   "tool without metadata",
			prompt: "Continue",
			history: []contracts.Message{
				{
					Role:       contracts.MessageRoleTool,
					Content:    "Some result",
					ToolCallID: "tool_456",
				},
			},
			expected: "Tool unknown result: Some result\nUser: Continue",
		},
		{
			name:   "assistant message with empty content",
			prompt: "Next",
			history: []contracts.Message{
				{Role: contracts.RoleAssistant, Content: ""},
				{Role: contracts.RoleUser, Content: "Previous"},
			},
			expected: "User: Previous\nUser: Next",
		},
		{
			name:   "basic conversation with clear role markers",
			prompt: "What can you help me with?",
			history: []contracts.Message{
				{Role: contracts.RoleUser, Content: "Hello, how are you?"},
				{Role: contracts.RoleAssistant, Content: "I'm doing well, thank you!"},
			},
			expected: "User: Hello, how are you?\nAssistant: I'm doing well, thank you!\nUser: What can you help me with?",
		},
		{
			name:   "conversation with tool messages",
			prompt: "What's the cluster status?",
			history: []contracts.Message{
				{Role: contracts.RoleUser, Content: "list which clusters I have available"},
				{Role: contracts.RoleAssistant, Content: `{"reasoning":["User is requesting a list of available clusters"]}`},
				{
					Role:       contracts.MessageRoleTool,
					Content:    `{"query": "list all EKS clusters", "output": "eks-cluster-1"}`,
					ToolCallID: "tool_789",
					Metadata:   map[string]interface{}{"tool_name": "cluster_list"},
				},
				{Role: contracts.RoleAssistant, Content: "You have eks-cluster-1 available"},
			},
			expected: "User: list which clusters I have available\nAssistant: {\"reasoning\":[\"User is requesting a list of available clusters\"]}\nTool cluster_list result: {\"query\": \"list all EKS clusters\", \"output\": \"eks-cluster-1\"}\nAssistant: You have eks-cluster-1 available\nUser: What's the cluster status?",
		},
		{
			name:   "single message",
			prompt: "How are you?",
			history: []contracts.Message{
				{Role: contracts.RoleUser, Content: "Hello"},
			},
			expected: "User: Hello\nUser: How are you?",
		},
		{
			name:   "system message included with conversation",
			prompt: "What should I do next?",
			history: []contracts.Message{
				{Role: contracts.MessageRoleSystem, Content: "You are a helpful assistant"},
				{Role: contracts.RoleUser, Content: "Hi"},
				{Role: contracts.RoleAssistant, Content: "Hello!"},
			},
			expected: "System: You are a helpful assistant\nUser: Hi\nAssistant: Hello!\nUser: What should I do next?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock memory, but pass nil for "no memory" test case
			var memory contracts.Memory
			if tt.history != nil {
				memory = &mockMemory{messages: tt.history}
			}

			result := BuildInlineHistoryPrompt(context.Background(), tt.prompt, memory, nil)
			if result != tt.expected {
				t.Errorf("BuildInlineHistoryPrompt() mismatch\nGot:\n%s\n\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}

// mockMemory is a simple in-memory implementation for testing
type mockMemory struct {
	messages []contracts.Message
}

func (m *mockMemory) AddMessage(ctx context.Context, message contracts.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *mockMemory) GetMessages(ctx context.Context, options ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	return m.messages, nil
}

func (m *mockMemory) Clear(ctx context.Context) error {
	m.messages = nil
	return nil
}
