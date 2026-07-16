package gemini

import (
	"bytes"
	"context"
	"testing"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

func TestGeminiMessageHistoryBuilder_BuildContents(t *testing.T) {
	logger := telemetry.NewLogger()
	builder := geminiNewMessageHistoryBuilder(logger)

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
				Memory: &geminiMockMemory{
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
				Memory: &geminiMockMemory{
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
				Memory: &geminiMockMemory{
					messages: []contracts.Message{
						{Role: contracts.RoleUser, Content: "Get weather"},
						{Role: contracts.RoleAssistant, Content: "I'll check the weather", ToolCalls: []contracts.ToolCall{
							{ID: "call_123", Name: "get_weather", Arguments: `{"location": "NYC"}`},
						}},
						{Role: contracts.MessageRoleTool, Content: "Sunny, 72°F", ToolCallID: "call_123", Metadata: map[string]interface{}{
							"tool_name": "get_weather",
						}},
						{Role: contracts.RoleUser, Content: "What's next?"}, // Agent adds current prompt to memory by default
					},
				},
			},
			expected: 4, // 3 from memory + 1 current prompt
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents := builder.buildContents(context.Background(), tt.prompt, tt.params)
			if len(contents) != tt.expected {
				t.Errorf("Expected %d contents, got %d", tt.expected, len(contents))
			}
		})
	}
}

func TestGeminiConvertMemoryMessage_PreservesThoughtSignature(t *testing.T) {
	logger := telemetry.NewLogger()
	builder := geminiNewMessageHistoryBuilder(logger)

	sig := []byte("test-thought-signature-bytes")

	msg := contracts.Message{
		Role:    contracts.RoleAssistant,
		Content: "",
		ToolCalls: []contracts.ToolCall{
			{
				ID:               "call_1",
				Name:             "get_weather",
				Arguments:        `{"location":"NYC"}`,
				ThoughtSignature: sig,
			},
		},
	}

	content := builder.convertMemoryMessage(msg)
	if content == nil {
		t.Fatal("expected non-nil content for assistant message with tool calls")
	}
	if content.Role != "model" {
		t.Fatalf("expected role 'model', got %q", content.Role)
	}
	if len(content.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(content.Parts))
	}

	part := content.Parts[0]
	if part.FunctionCall == nil {
		t.Fatal("expected FunctionCall to be set")
	}
	if part.FunctionCall.Name != "get_weather" {
		t.Errorf("expected function name 'get_weather', got %q", part.FunctionCall.Name)
	}
	if !bytes.Equal(part.ThoughtSignature, sig) {
		t.Errorf("ThoughtSignature not preserved: got %v, want %v", part.ThoughtSignature, sig)
	}
}

func TestGeminiConvertMemoryMessage_NilSignatureOmitted(t *testing.T) {
	logger := telemetry.NewLogger()
	builder := geminiNewMessageHistoryBuilder(logger)

	msg := contracts.Message{
		Role: contracts.RoleAssistant,
		ToolCalls: []contracts.ToolCall{
			{ID: "call_1", Name: "search", Arguments: `{}`},
		},
	}

	content := builder.convertMemoryMessage(msg)
	if content == nil {
		t.Fatal("expected non-nil content")
	}

	part := content.Parts[0]
	if part.ThoughtSignature != nil {
		t.Errorf("expected nil ThoughtSignature for tool call without signature, got %v", part.ThoughtSignature)
	}
}
