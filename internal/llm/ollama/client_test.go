package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nu/internal/contracts"
	"nu/internal/llm"
	"nu/internal/telemetry"
)

func TestOllamaNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:11434", client.BaseURL)
	assert.Equal(t, "qwen3:0.6b", client.Model)
	assert.NotNil(t, client.HTTPClient)
	assert.NotNil(t, client.logger)
}

func TestOllamaNewClientWithOptions(t *testing.T) {
	logger := telemetry.NewLogger()
	client := NewClient(
		WithModel("mistral"),
		WithBaseURL("http://localhost:8080"),
		WithLogger(logger),
	)

	assert.Equal(t, "mistral", client.Model)
	assert.Equal(t, "http://localhost:8080", client.BaseURL)
	assert.Equal(t, logger, client.logger)
}

func TestOllamaGenerate(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var req GenerateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-model", req.Model)
		assert.Equal(t, "User: test prompt", req.Prompt)
		assert.False(t, req.Stream)
		assert.Equal(t, 0.8, req.Options.Temperature)

		// Return response
		response := GenerateResponse{
			Model:    "test-model",
			Response: "This is a test response",
			Done:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Create client
	client := NewClient(
		WithModel("test-model"),
		WithBaseURL(server.URL),
	)

	// Test generate
	response, err := client.Generate(
		context.Background(),
		"test prompt",
		WithTemperature(0.8),
	)

	require.NoError(t, err)
	assert.Equal(t, "This is a test response", response)
}

func TestOllamaGenerateWithSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GenerateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "You are a helpful assistant", req.System)

		response := GenerateResponse{
			Model:    "test-model",
			Response: "System message received",
			Done:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(
		WithModel("test-model"),
		WithBaseURL(server.URL),
	)

	response, err := client.Generate(
		context.Background(),
		"test prompt",
		WithSystemMessage("You are a helpful assistant"),
	)

	require.NoError(t, err)
	assert.Equal(t, "System message received", response)
}

func TestOllamaChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/chat", r.URL.Path)

		var req ChatRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-model", req.Model)
		assert.Len(t, req.Messages, 2)
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "user", req.Messages[1].Role)

		response := ChatResponse{
			Model: "test-model",
			Message: ChatMessage{
				Role:    "assistant",
				Content: "Hello! How can I help you?",
			},
			Done: true,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(
		WithModel("test-model"),
		WithBaseURL(server.URL),
	)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant",
		},
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	response, err := client.Chat(context.Background(), messages, &llm.GenerateParams{
		Temperature: 0.7,
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello! How can I help you?", response)
}

func TestOllamaGenerateWithTools(t *testing.T) {
	// First request: model returns tool_calls. Second request (after we
	// feed back the tool result): model returns the final answer with no
	// further tool calls.
	turn := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/chat", r.URL.Path)

		var req ChatRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Tool definitions must be sent on every request.
		require.Len(t, req.Tools, 1)
		assert.Equal(t, "function", req.Tools[0].Type)
		assert.Equal(t, "test-tool", req.Tools[0].Function.Name)
		assert.Equal(t, "A test tool", req.Tools[0].Function.Description)

		var resp ChatResponse
		switch turn {
		case 0:
			// First call: model decides to invoke the tool.
			resp = ChatResponse{
				Model: "test-model",
				Message: ChatMessage{
					Role: "assistant",
					ToolCalls: []ToolCall{{
						Function: ToolCallFunction{
							Name:      "test-tool",
							Arguments: map[string]interface{}{"q": "hello"},
						},
					}},
				},
				Done: true,
			}
		case 1:
			// Second call: conversation now includes the tool result.
			require.GreaterOrEqual(t, len(req.Messages), 3)
			assert.Equal(t, "tool", req.Messages[len(req.Messages)-1].Role)
			assert.Equal(t, "tool result", req.Messages[len(req.Messages)-1].Content)
			resp = ChatResponse{
				Model: "test-model",
				Message: ChatMessage{
					Role:    "assistant",
					Content: "I can help you with that using the available tools",
				},
				Done: true,
			}
		default:
			t.Fatalf("unexpected extra request, turn=%d", turn)
		}
		turn++

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(
		WithModel("test-model"),
		WithBaseURL(server.URL),
	)

	ollamaMockTool := &ollamaMockTool{
		name:        "test-tool",
		description: "A test tool",
		runResult:   "tool result",
	}

	response, err := client.GenerateWithTools(
		context.Background(),
		"Help me with something",
		[]contracts.Tool{ollamaMockTool},
	)

	require.NoError(t, err)
	assert.Equal(t, "I can help you with that using the available tools", response)
	assert.Equal(t, 2, turn, "expected exactly two roundtrips")
}

// TestGenerateWithTools_NoToolCalls covers the happy path where the model
// responds with a final answer on the first turn — no loop iterations.
func TestOllamaGenerateWithTools_NoToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatResponse{
			Model:   "test-model",
			Message: ChatMessage{Role: "assistant", Content: "I can answer directly"},
			Done:    true,
		})
	}))
	defer server.Close()

	client := NewClient(WithModel("test-model"), WithBaseURL(server.URL))
	resp, err := client.GenerateWithTools(context.Background(), "hello",
		[]contracts.Tool{&ollamaMockTool{name: "noop", description: "noop"}})
	require.NoError(t, err)
	assert.Equal(t, "I can answer directly", resp)
}

// TestGenerateWithTools_MaxIterationsExceeded ensures we bound runaway
// loops when the model keeps requesting tools.
func TestOllamaGenerateWithTools_MaxIterationsExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatResponse{
			Model: "test-model",
			Message: ChatMessage{
				Role: "assistant",
				ToolCalls: []ToolCall{{
					Function: ToolCallFunction{Name: "noop", Arguments: map[string]interface{}{}},
				}},
			},
			Done: true,
		})
	}))
	defer server.Close()

	client := NewClient(WithModel("test-model"), WithBaseURL(server.URL))
	_, err := client.GenerateWithTools(context.Background(), "loop",
		[]contracts.Tool{&ollamaMockTool{name: "noop", description: "noop", runResult: "ok"}},
		contracts.WithMaxIterations(2),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations")
}

// TestGenerateWithTools_ToolNotFound feeds an error message back to the
// model when it hallucinates a tool name and continues the loop instead
// of failing the whole call.
func TestOllamaGenerateWithTools_ToolNotFound(t *testing.T) {
	turn := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		var resp ChatResponse
		switch turn {
		case 0:
			resp = ChatResponse{Message: ChatMessage{
				Role: "assistant",
				ToolCalls: []ToolCall{{
					Function: ToolCallFunction{Name: "ghost_tool", Arguments: map[string]interface{}{}},
				}},
			}, Done: true}
		case 1:
			require.GreaterOrEqual(t, len(req.Messages), 2)
			last := req.Messages[len(req.Messages)-1]
			assert.Equal(t, "tool", last.Role)
			assert.Contains(t, last.Content, "ghost_tool")
			assert.Contains(t, last.Content, "not found")
			resp = ChatResponse{Message: ChatMessage{Role: "assistant", Content: "I'll skip that"}, Done: true}
		}
		turn++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithModel("test-model"), WithBaseURL(server.URL))
	resp, err := client.GenerateWithTools(context.Background(), "hi",
		[]contracts.Tool{&ollamaMockTool{name: "real_tool", description: "the real one"}})
	require.NoError(t, err)
	assert.Equal(t, "I'll skip that", resp)
}

// TestGenerateWithTools_PersistsToolExchangesToMemory is the regression
// test for the #325 review BLOCKER: across multi-turn conversations the
// tool exchange must end up in Memory so the next turn can replay it via
// BuildInlineHistoryPrompt.
func TestOllamaGenerateWithTools_PersistsToolExchangesToMemory(t *testing.T) {
	turn := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		var resp ChatResponse
		switch turn {
		case 0:
			resp = ChatResponse{Message: ChatMessage{
				Role: "assistant",
				ToolCalls: []ToolCall{{
					Function: ToolCallFunction{Name: "calc", Arguments: map[string]interface{}{"expr": "1+1"}},
				}},
			}, Done: true}
		case 1:
			resp = ChatResponse{Message: ChatMessage{Role: "assistant", Content: "the answer is 2"}, Done: true}
		}
		turn++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	mem := ollamaNewRecordingMemory()
	client := NewClient(WithModel("test-model"), WithBaseURL(server.URL))
	_, err := client.GenerateWithTools(context.Background(), "what is 1+1",
		[]contracts.Tool{&ollamaMockTool{name: "calc", description: "calculator", runResult: "2"}},
		contracts.WithMemory(mem),
	)
	require.NoError(t, err)

	require.Len(t, mem.added, 2, "expected assistant tool-call message + tool result message")

	asst := mem.added[0]
	assert.Equal(t, contracts.RoleAssistant, asst.Role)
	require.Len(t, asst.ToolCalls, 1)
	assert.Equal(t, "calc", asst.ToolCalls[0].Name)
	assert.Contains(t, asst.ToolCalls[0].Arguments, `"expr":"1+1"`)

	tool := mem.added[1]
	assert.Equal(t, contracts.MessageRoleTool, tool.Role)
	assert.Equal(t, "2", tool.Content)
	assert.NotEmpty(t, tool.ToolCallID, "ToolCallID required for BuildInlineHistoryPrompt to render the tool message back")
	assert.Equal(t, "calc", tool.Metadata["tool_name"])
}

// TestGenerateWithTools_ParallelCallsGetUniqueIDs covers the case where
// the model invokes the same tool twice in a single assistant turn. Each
// invocation must get a unique synthesized ToolCallID; otherwise the
// memory pairing between assistant tool_call and tool result is broken.
func TestOllamaGenerateWithTools_ParallelCallsGetUniqueIDs(t *testing.T) {
	turn := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		var resp ChatResponse
		switch turn {
		case 0:
			// Same tool invoked twice in a single assistant message.
			resp = ChatResponse{Message: ChatMessage{
				Role: "assistant",
				ToolCalls: []ToolCall{
					{Function: ToolCallFunction{Name: "calc", Arguments: map[string]interface{}{"expr": "1+1"}}},
					{Function: ToolCallFunction{Name: "calc", Arguments: map[string]interface{}{"expr": "2+2"}}},
				},
			}, Done: true}
		case 1:
			resp = ChatResponse{Message: ChatMessage{Role: "assistant", Content: "done"}, Done: true}
		}
		turn++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	mem := ollamaNewRecordingMemory()
	client := NewClient(WithModel("test-model"), WithBaseURL(server.URL))
	_, err := client.GenerateWithTools(context.Background(), "do both",
		[]contracts.Tool{&ollamaMockTool{name: "calc", description: "calculator", runResult: "ok"}},
		contracts.WithMemory(mem),
	)
	require.NoError(t, err)

	// One assistant message + two tool result messages
	require.Len(t, mem.added, 3)
	asst := mem.added[0]
	require.Len(t, asst.ToolCalls, 2)
	assert.NotEqual(t, asst.ToolCalls[0].ID, asst.ToolCalls[1].ID,
		"parallel calls must get unique synthesized IDs")

	tool1, tool2 := mem.added[1], mem.added[2]
	assert.Equal(t, asst.ToolCalls[0].ID, tool1.ToolCallID,
		"first tool result ID must match first assistant tool_call ID")
	assert.Equal(t, asst.ToolCalls[1].ID, tool2.ToolCallID,
		"second tool result ID must match second assistant tool_call ID")
	assert.NotEqual(t, tool1.ToolCallID, tool2.ToolCallID)
}

// ollamaRecordingMemory is a minimal in-memory Memory that captures every
// message added to it for assertion in tests.
type ollamaRecordingMemory struct {
	added []contracts.Message
}

func ollamaNewRecordingMemory() *ollamaRecordingMemory { return &ollamaRecordingMemory{} }

func (m *ollamaRecordingMemory) AddMessage(_ context.Context, msg contracts.Message) error {
	m.added = append(m.added, msg)
	return nil
}

func (m *ollamaRecordingMemory) GetMessages(_ context.Context, _ ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	return m.added, nil
}

func (m *ollamaRecordingMemory) Clear(_ context.Context) error {
	m.added = nil
	return nil
}

func TestOllamaListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/tags", r.URL.Path)

		response := struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}{
			Models: []struct {
				Name string `json:"name"`
			}{
				{Name: "llama2"},
				{Name: "mistral"},
				{Name: "codellama"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	models, err := client.ListModels(context.Background())

	require.NoError(t, err)
	assert.Len(t, models, 3)
	assert.Contains(t, models, "llama2")
	assert.Contains(t, models, "mistral")
	assert.Contains(t, models, "codellama")
}

func TestOllamaPullModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/pull", r.URL.Path)

		var req struct {
			Name string `json:"name"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "mistral", req.Name)

		response := struct {
			Status string `json:"status"`
		}{
			Status: "success",
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	err := client.PullModel(context.Background(), "mistral")

	require.NoError(t, err)
}

func TestOllamaMakeRequestError(t *testing.T) {
	// Create a server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Internal Server Error"))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	_, err := client.Generate(context.Background(), "test prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500")
}

func TestOllamaName(t *testing.T) {
	client := NewClient()
	assert.Equal(t, "ollama", client.Name())
}

// Mock tool for testing
type ollamaMockTool struct {
	name        string
	description string
	runResult   string
}

func (t *ollamaMockTool) Name() string {
	return t.name
}

func (t *ollamaMockTool) DisplayName() string {
	return t.name
}

func (t *ollamaMockTool) Description() string {
	return t.description
}

func (t *ollamaMockTool) Internal() bool {
	return false
}

func (t *ollamaMockTool) Run(ctx context.Context, input string) (string, error) {
	if t.runResult != "" {
		return t.runResult, nil
	}
	return "mock result", nil
}

func (t *ollamaMockTool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"input": {
			Type:        "string",
			Description: "Input parameter",
			Required:    true,
		},
	}
}

func (t *ollamaMockTool) Execute(ctx context.Context, args string) (string, error) {
	if t.runResult != "" {
		return t.runResult, nil
	}
	return "mock result", nil
}

// Test GenerateOption functions
func TestOllamaWithTemperature(t *testing.T) {
	options := &contracts.GenerateOptions{}
	WithTemperature(0.5)(options)

	assert.NotNil(t, options.LLMConfig)
	assert.Equal(t, 0.5, options.LLMConfig.Temperature)
}

func TestOllamaWithTopP(t *testing.T) {
	options := &contracts.GenerateOptions{}
	WithTopP(0.9)(options)

	assert.NotNil(t, options.LLMConfig)
	assert.Equal(t, 0.9, options.LLMConfig.TopP)
}

func TestOllamaWithStopSequences(t *testing.T) {
	options := &contracts.GenerateOptions{}
	stopSequences := []string{"stop1", "stop2"}
	WithStopSequences(stopSequences)(options)

	assert.NotNil(t, options.LLMConfig)
	assert.Equal(t, stopSequences, options.LLMConfig.StopSequences)
}

func TestOllamaWithSystemMessage(t *testing.T) {
	options := &contracts.GenerateOptions{}
	WithSystemMessage("You are helpful")(options)

	assert.Equal(t, "You are helpful", options.SystemMessage)
}
