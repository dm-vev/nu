package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"nu/internal/contracts"
	"nu/internal/llm"
	provider "nu/internal/llm/openai"
	"nu/internal/telemetry"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func TestOpenAIGenerate(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header with test-key")
		}

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "test response",
						Role:    "assistant",
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create our wrapper client with a logger
	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)

	// Override the client to use our test server
	// We need to create a new client with the test server URL
	testClient := openai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)
	client.Client = testClient
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	// Test generation
	resp, err := client.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	if resp != "test response" {
		t.Errorf("Expected response 'test response', got '%s'", resp)
	}
}

func TestOpenAIGenerate_OmitsZeroTopPByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if _, ok := reqBody["top_p"]; ok {
			t.Fatalf("expected top_p to be omitted when no GenerateOptions are provided")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openai.ChatCompletion{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok", Role: "assistant"}}}})
	}))
	defer server.Close()

	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	if _, err := client.Generate(context.Background(), "who are you"); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
}

func TestOpenAIGenerate_IncludesTopPWhenExplicitlySet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		v, ok := reqBody["top_p"].(float64)
		if !ok {
			t.Fatalf("expected top_p in request when explicitly set")
		}
		if v != 0.9 {
			t.Fatalf("expected top_p=0.9, got %v", v)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openai.ChatCompletion{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok", Role: "assistant"}}}})
	}))
	defer server.Close()

	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	if _, err := client.Generate(context.Background(), "who are you", provider.WithTopP(0.9)); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
}

func TestOpenAIChat(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "test response",
						Role:    "assistant",
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create our wrapper client with a logger
	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)

	// Override the client to use our test server
	testClient := openai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)
	client.Client = testClient
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	// Test chat
	messages := []llm.Message{
		{
			Role:    "user",
			Content: "test message",
		},
	}

	resp, err := client.Chat(context.Background(), messages, nil)
	if err != nil {
		t.Fatalf("Failed to chat: %v", err)
	}

	if resp != "test response" {
		t.Errorf("Expected response 'test response', got '%s'", resp)
	}
}

func TestOpenAIGenerateWithResponseFormat(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify response format is present
		if reqBody["response_format"] == nil {
			t.Error("Expected response_format in request")
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: `{"name": "test", "value": 123}`,
						Role:    "assistant",
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create our wrapper client with a logger
	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)

	// Override the client to use our test server
	testClient := openai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)
	client.Client = testClient
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	// Test generation with response format
	resp, err := client.Generate(context.Background(), "test prompt",
		provider.WithResponseFormat(contracts.ResponseFormat{
			Name: "test_format",
			Schema: contracts.JSONSchema{
				"type": "object",
				"properties": map[string]interface{}{
					"name":  map[string]interface{}{"type": "string"},
					"value": map[string]interface{}{"type": "number"},
				},
			},
		}),
	)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	expected := `{"name": "test", "value": 123}`
	if resp != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, resp)
	}
}

func TestOpenAIChatWithToolMessages(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify that tool messages are present with tool_call_id
		messages := reqBody["messages"].([]interface{})
		foundToolMessage := false
		for _, msg := range messages {
			msgMap := msg.(map[string]interface{})
			if msgMap["role"] == "tool" {
				foundToolMessage = true
				if msgMap["tool_call_id"] != "test-tool-call-id" {
					t.Errorf("Expected tool_call_id 'test-tool-call-id', got '%s'", msgMap["tool_call_id"])
				}
				break
			}
		}
		if !foundToolMessage {
			t.Error("Expected tool message in request")
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "test response",
						Role:    "assistant",
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create our wrapper client with a logger
	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)

	// Override the client to use our test server
	testClient := openai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)
	client.Client = testClient
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	// Test chat with tool messages
	messages := []llm.Message{
		{
			Role:    "user",
			Content: "test message",
		},
		{
			Role:       "tool",
			Content:    "tool result",
			ToolCallID: "test-tool-call-id",
		},
	}

	resp, err := client.Chat(context.Background(), messages, nil)
	if err != nil {
		t.Fatalf("Failed to chat: %v", err)
	}

	if resp != "test response" {
		t.Errorf("Expected response 'test response', got '%s'", resp)
	}
}

func TestOpenAIParallelToolExecution(t *testing.T) {
	// Create a test server that simulates parallel tool calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Check if this is the first request (with tools) or second request (with tool results)
		messages := reqBody["messages"].([]interface{})
		hasToolResults := false
		for _, msg := range messages {
			msgMap := msg.(map[string]interface{})
			if msgMap["role"] == "tool" {
				hasToolResults = true
				break
			}
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		var response openai.ChatCompletion

		if !hasToolResults {
			// First request - return tool calls
			response = openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "",
							Role:    "assistant",
							ToolCalls: []openai.ChatCompletionMessageToolCallUnion{
								{
									ID: "call_123",
									Function: openai.ChatCompletionMessageFunctionToolCallFunction{
										Name: "parallel_tool_use",
										Arguments: `{
											"tool_uses": [
												{
													"recipient_name": "test_tool_1",
													"parameters": {"param1": "value1"}
												},
												{
													"recipient_name": "test_tool_2",
													"parameters": {"param2": "value2"}
												}
											]
										}`,
									},
								},
							},
						},
					},
				},
			}
		} else {
			// Second request - return final response
			response = openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Final response after parallel tools",
							Role:    "assistant",
						},
					},
				},
			}
		}

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create our wrapper client with a logger
	logger := telemetry.NewLogger()
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-4"),
		provider.WithLogger(logger),
	)

	// Override the client to use our test server
	testClient := openai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)
	client.Client = testClient
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	// Create mock tools
	mockTools := []contracts.Tool{
		&openAIMockTool{name: "test_tool_1", description: "Test tool 1"},
		&openAIMockTool{name: "test_tool_2", description: "Test tool 2"},
	}

	// Test parallel tool execution
	resp, err := client.GenerateWithTools(context.Background(), "test prompt", mockTools)
	if err != nil {
		t.Fatalf("Failed to generate with tools: %v", err)
	}

	expected := "Final response after parallel tools"
	if resp != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, resp)
	}
}

// openAIMockTool implements contracts.Tool for testing
type openAIMockTool struct {
	name        string
	description string
}

func (m *openAIMockTool) Name() string {
	return m.name
}

func (m *openAIMockTool) DisplayName() string {
	return m.name
}

func (m *openAIMockTool) Description() string {
	return m.description
}

func (m *openAIMockTool) Internal() bool {
	return false
}

func (m *openAIMockTool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"param": {
			Type:        "string",
			Description: "Test parameter",
			Required:    true,
		},
	}
}

func (m *openAIMockTool) Execute(ctx context.Context, args string) (string, error) {
	return fmt.Sprintf("Result from %s: %s", m.name, args), nil
}

func (m *openAIMockTool) Run(ctx context.Context, input string) (string, error) {
	return m.Execute(ctx, input)
}

func TestOpenAIReasoningEffort(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify reasoning_effort is present
		if reqBody["reasoning_effort"] != "low" {
			t.Errorf("Expected reasoning_effort 'low', got '%v'", reqBody["reasoning_effort"])
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "test", Role: "assistant"}},
			},
		})
	}))
	defer server.Close()

	// Create client
	client := provider.NewClient("test-key",
		provider.WithModel("gpt-5-mini"),
		provider.WithLogger(telemetry.NewLogger()),
	)
	client.ChatService = openai.NewChatService(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	// Test with reasoning effort
	_, err := client.Generate(context.Background(), "test",
		provider.WithReasoning("low"),
	)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
}

// openAIMockMemory is a simple in-memory implementation for testing
type openAIMockMemory struct {
	messages []contracts.Message
}

func (m *openAIMockMemory) AddMessage(ctx context.Context, message contracts.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *openAIMockMemory) GetMessages(ctx context.Context, options ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	return m.messages, nil
}

func (m *openAIMockMemory) Clear(ctx context.Context) error {
	m.messages = nil
	return nil
}

func TestOpenAIGenerateWithMemory(t *testing.T) {
	tests := []struct {
		name     string
		history  []contracts.Message
		prompt   string
		expected int // expected number of messages in request
	}{
		{
			name:     "empty memory",
			history:  nil, // No memory provided
			prompt:   "Hello",
			expected: 1, // Just the current user message
		},
		{
			name: "conversation with system message",
			history: []contracts.Message{
				{Role: contracts.MessageRoleSystem, Content: "You are helpful"},
				{Role: contracts.RoleUser, Content: "Hi"},
				{Role: contracts.RoleAssistant, Content: "Hello!"},
				{Role: contracts.RoleUser, Content: "How are you?"}, // Current prompt should be in memory
			},
			prompt:   "How are you?",
			expected: 4, // system + user + assistant + current user (from memory)
		},
		{
			name: "conversation with tool call",
			history: []contracts.Message{
				{Role: contracts.RoleUser, Content: "Check status"},
				{Role: contracts.RoleAssistant, Content: "Checking..."},
				{
					Role:       contracts.MessageRoleTool,
					Content:    "All good",
					ToolCallID: "call_123",
					Metadata:   map[string]interface{}{"tool_name": "status_check"},
				},
				{Role: contracts.RoleUser, Content: "Thanks"}, // Current prompt should be in memory
			},
			prompt:   "Thanks",
			expected: 4, // user + assistant + tool + current user (from memory)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that validates the request structure
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Parse request body to validate messages
				var reqBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				// Verify messages array
				messages, ok := reqBody["messages"].([]interface{})
				if !ok {
					t.Fatalf("Expected messages array in request")
				}

				if len(messages) != tt.expected {
					t.Errorf("Expected %d messages in request, got %d", tt.expected, len(messages))
				}

				// Verify system message comes first if present
				if len(tt.history) > 0 {
					hasSystemMessage := false
					for _, msg := range tt.history {
						if msg.Role == contracts.MessageRoleSystem {
							hasSystemMessage = true
							break
						}
					}

					if hasSystemMessage && len(messages) > 0 {
						firstMsg := messages[0].(map[string]interface{})
						if firstMsg["role"] != "system" {
							t.Errorf("Expected first message to be system message, got: %v", firstMsg["role"])
						}
					}
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				response := openai.ChatCompletion{
					Choices: []openai.ChatCompletionChoice{
						{
							Message: openai.ChatCompletionMessage{
								Content: "test response",
								Role:    "assistant",
							},
						},
					},
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Fatalf("Failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			// Create client with test server
			client := provider.NewClient("test-key",
				provider.WithBaseURL(server.URL),
				provider.WithLogger(telemetry.NewLogger()))

			var memory contracts.Memory
			if tt.history != nil {
				memory = &openAIMockMemory{messages: tt.history}
			}

			// Test Generate with memory
			_, err := client.Generate(context.Background(), tt.prompt,
				contracts.WithMemory(memory))

			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
		})
	}
}
