package http

import (
	"net/http"

	"nu/internal/agent"
)

// Server provides HTTP/SSE endpoints for agent streaming
type Server struct {
	Agent  *agent.Agent
	Port   int
	Server *http.Server
}

// StreamRequest represents the JSON request for streaming.
type StreamRequest struct {
	Input          string            `json:"input"`
	OrgID          string            `json:"org_id,omitempty"`
	ConversationID string            `json:"conversation_id,omitempty"`
	Context        map[string]string `json:"context,omitempty"`
	MaxIterations  int               `json:"max_iterations,omitempty"`
}

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	ID        string      `json:"id,omitempty"`
	Retry     int         `json:"retry,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// StreamEventData represents data for streaming events.
type StreamEventData struct {
	Type         string                 `json:"type"`
	Content      string                 `json:"content,omitempty"`
	ThinkingStep string                 `json:"thinking_step,omitempty"`
	ToolCall     *ToolCallData          `json:"tool_call,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IsFinal      bool                   `json:"is_final"`
	Timestamp    int64                  `json:"timestamp"`
}

// ToolCallData represents tool call information for HTTP/SSE.
type ToolCallData struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
	Result    string `json:"result,omitempty"`
	Status    string `json:"status"`
}
