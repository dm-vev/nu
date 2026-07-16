package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nu/internal/agent"
	"nu/internal/contracts"
	"nu/internal/memory"
)

// MockLLM implements a simple mock LLM for testing
type MockLLM struct {
	response string
	err      error
}

func (m *MockLLM) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockLLM) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockLLM) Name() string {
	return "mock-llm"
}

func (m *MockLLM) SupportsStreaming() bool {
	return true
}

func (m *MockLLM) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	content, err := m.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}
	return &contracts.LLMResponse{
		Content:    content,
		Model:      "mock-llm",
		StopReason: "complete",
		Usage: &contracts.TokenUsage{
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
		},
		Metadata: map[string]interface{}{
			"provider": "mock",
		},
	}, nil
}

func (m *MockLLM) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	content, err := m.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}
	return &contracts.LLMResponse{
		Content:    content,
		Model:      "mock-llm",
		StopReason: "complete",
		Usage: &contracts.TokenUsage{
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
		},
		Metadata: map[string]interface{}{
			"provider":   "mock",
			"tools_used": true,
		},
	}, nil
}

// Implement streaming methods
func (m *MockLLM) GenerateStream(ctx context.Context, prompt string, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	eventChan := make(chan contracts.StreamEvent, 10)

	go func() {
		defer close(eventChan)

		if m.err != nil {
			eventChan <- contracts.StreamEvent{
				Type:      contracts.StreamEventError,
				Error:     m.err,
				Timestamp: time.Now(),
			}
			return
		}

		// Send message start
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStart,
			Timestamp: time.Now(),
		}

		// Send content in chunks
		words := strings.Split(m.response, " ")
		for _, word := range words {
			eventChan <- contracts.StreamEvent{
				Type:      contracts.StreamEventContentDelta,
				Content:   word + " ",
				Timestamp: time.Now(),
			}

			// Small delay to simulate real streaming
			time.Sleep(1 * time.Millisecond)
		}

		// Send completion
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStop,
			Timestamp: time.Now(),
		}
	}()

	return eventChan, nil
}

func (m *MockLLM) GenerateWithToolsStream(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	return m.GenerateStream(ctx, prompt, options...)
}

// MockStreamingAgent wraps MockLLM to provide agent streaming
type MockStreamingAgent struct {
	*agent.Agent
	llm *MockLLM
}

func (m *MockStreamingAgent) RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	eventChan := make(chan contracts.AgentStreamEvent, 10)

	go func() {
		defer close(eventChan)

		// Get LLM stream
		llmEventChan, err := m.llm.GenerateStream(ctx, input)
		if err != nil {
			eventChan <- contracts.AgentStreamEvent{
				Type:      contracts.AgentEventError,
				Error:     err,
				Timestamp: time.Now(),
			}
			return
		}

		// Forward LLM events as agent events
		for llmEvent := range llmEventChan {
			var agentEventType contracts.AgentEventType
			switch llmEvent.Type {
			case contracts.StreamEventContentDelta:
				agentEventType = contracts.AgentEventContent
			case contracts.StreamEventError:
				agentEventType = contracts.AgentEventError
			default:
				agentEventType = contracts.AgentEventContent
			}

			eventChan <- contracts.AgentStreamEvent{
				Type:      agentEventType,
				Content:   llmEvent.Content,
				Error:     llmEvent.Error,
				Timestamp: llmEvent.Timestamp,
			}
		}

		// Send completion
		eventChan <- contracts.AgentStreamEvent{
			Type:      contracts.AgentEventComplete,
			Timestamp: time.Now(),
		}
	}()

	return eventChan, nil
}

func createTestAgent(response string, err error) contracts.StreamingAgent {
	mockLLM := &MockLLM{response: response, err: err}
	memoryStore := memory.NewConversationBuffer()

	agentInstance, _ := agent.NewAgent(
		agent.WithLLM(mockLLM),
		agent.WithMemory(memoryStore),
		agent.WithName("TestAgent"),
		agent.WithOrgID("test-org"), // Add org ID for memory operations
	)

	return &MockStreamingAgent{
		Agent: agentInstance,
		llm:   mockLLM,
	}
}

func TestHTTPServer_Health(t *testing.T) {
	// Create test agent
	testAgent := createTestAgent("test response", nil)

	// Create HTTP server
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Test health endpoint
	server.HandleHealth(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	if response["agent"] != "TestAgent" {
		t.Errorf("Expected agent 'TestAgent', got %v", response["agent"])
	}
}

func TestHTTPServer_Metadata(t *testing.T) {
	// Create test agent
	testAgent := createTestAgent("test response", nil)

	// Create HTTP server
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/agent/metadata", nil)
	w := httptest.NewRecorder()

	// Test metadata endpoint
	server.HandleMetadata(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["name"] != "TestAgent" {
		t.Errorf("Expected name 'TestAgent', got %v", response["name"])
	}

	if response["supports_streaming"] != true {
		t.Errorf("Expected supports_streaming true, got %v", response["supports_streaming"])
	}
}

func TestHTTPServer_Run(t *testing.T) {
	// Create test agent
	testAgent := createTestAgent("Hello, world!", nil)

	// Create HTTP server
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Create test request
	requestData := StreamRequest{
		Input:          "test prompt",
		OrgID:          "test-org",
		ConversationID: "test-conversation",
	}
	requestBody, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/api/v1/agent/run", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test run endpoint
	server.HandleRun(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["output"] != "Hello, world!" {
		t.Errorf("Expected output 'Hello, world!', got %v", response["output"])
	}
}

func TestHTTPServer_Stream(t *testing.T) {
	// Create test agent
	testAgent := createTestAgent("Hello streaming world", nil)

	// Create HTTP server
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Create test request
	requestData := StreamRequest{
		Input:          "test streaming prompt",
		OrgID:          "test-org",
		ConversationID: "test-conversation",
	}
	requestBody, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/api/v1/agent/stream", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test stream endpoint
	server.HandleStream(w, req)

	// Check response headers
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got %v", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got %v", w.Header().Get("Cache-Control"))
	}

	// Check that we got SSE events
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "event: connected") {
		t.Error("Expected 'event: connected' in response")
	}

	if !strings.Contains(responseBody, "event: content") {
		t.Error("Expected 'event: content' in response")
	}

	if !strings.Contains(responseBody, "event: done") {
		t.Error("Expected 'event: done' in response")
	}

	// Check that content is properly formatted as SSE
	lines := strings.Split(responseBody, "\n")
	hasDataLine := false
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			hasDataLine = true
			// Try to parse the JSON data
			jsonData := strings.TrimPrefix(line, "data: ")
			var eventData StreamEventData
			if err := json.Unmarshal([]byte(jsonData), &eventData); err != nil {
				t.Errorf("Failed to parse SSE data as JSON: %v", err)
			}
		}
	}

	if !hasDataLine {
		t.Error("Expected at least one 'data: ' line in SSE response")
	}
}

func TestHTTPServer_StreamWithError(t *testing.T) {
	// Create test agent with error
	testAgent := createTestAgent("", &LLMError{Message: "test error"})

	// Create HTTP server
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Create test request
	requestData := StreamRequest{
		Input:          "test prompt",
		OrgID:          "test-org",
		ConversationID: "test-conversation",
	}
	requestBody, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/api/v1/agent/stream", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test stream endpoint
	server.HandleStream(w, req)

	// Check that we got an error event
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "event: error") {
		t.Error("Expected 'event: error' in response")
	}

	if !strings.Contains(responseBody, "test error") {
		t.Error("Expected error message 'test error' in response")
	}
}

func TestHTTPServer_CORS(t *testing.T) {
	// Create test agent
	testAgent := createTestAgent("test response", nil)

	// Create HTTP server
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Create a handler with CORS
	handler := server.AddCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test OPTIONS request (preflight)
	req := httptest.NewRequest("OPTIONS", "/api/v1/agent/stream", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got %v", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if !strings.Contains(w.Header().Get("Access-Control-Allow-Methods"), "POST") {
		t.Error("Expected 'POST' in Access-Control-Allow-Methods")
	}

	if !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Error("Expected 'Content-Type' in Access-Control-Allow-Headers")
	}
}

func TestStreamEventData_Conversion(t *testing.T) {
	// Create test agent event
	agentEvent := contracts.AgentStreamEvent{
		Type:         contracts.AgentEventContent,
		Content:      "test content",
		ThinkingStep: "test thinking",
		ToolCall: &contracts.ToolCallEvent{
			ID:        "call-123",
			Name:      "test-tool",
			Arguments: `{"arg": "value"}`,
			Result:    "test result",
			Status:    "completed",
		},
		Timestamp: time.Now(),
	}

	// Create HTTP server
	testAgent := createTestAgent("test", nil)
	server := NewServer(testAgent.(*MockStreamingAgent).Agent, 8080)

	// Convert event
	httpEvent := server.convertAgentEventToHTTPEvent(agentEvent)

	// Verify conversion
	if httpEvent.Type != string(contracts.AgentEventContent) {
		t.Errorf("Expected type %s, got %s", contracts.AgentEventContent, httpEvent.Type)
	}

	if httpEvent.Content != "test content" {
		t.Errorf("Expected content 'test content', got %s", httpEvent.Content)
	}

	if httpEvent.ThinkingStep != "test thinking" {
		t.Errorf("Expected thinking step 'test thinking', got %s", httpEvent.ThinkingStep)
	}

	if httpEvent.ToolCall == nil {
		t.Error("Expected tool call to be converted")
	} else {
		if httpEvent.ToolCall.ID != "call-123" {
			t.Errorf("Expected tool call ID 'call-123', got %s", httpEvent.ToolCall.ID)
		}
		if httpEvent.ToolCall.Name != "test-tool" {
			t.Errorf("Expected tool call name 'test-tool', got %s", httpEvent.ToolCall.Name)
		}
	}
}

// LLMError implements a simple error type for testing
type LLMError struct {
	Message string
}

func (e *LLMError) Error() string {
	return e.Message
}
