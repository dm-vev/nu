package telemetry

import (
	"context"
	"time"
)

// Message represents a chat message with role and content.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolCall represents a tool call made by the LLM.
type ToolCall struct {
	Name       string        `json:"name"`
	Arguments  string        `json:"arguments"`
	ID         string        `json:"id,omitempty"`
	Result     string        `json:"result,omitempty"`
	Error      string        `json:"error,omitempty"`
	Timestamp  string        `json:"timestamp"`
	DurationMs int64         `json:"duration_ms,omitempty"`
	StartTime  time.Time     `json:"-"`
	Duration   time.Duration `json:"-"`
}

type toolCallsKey struct{}

// WithToolCallsCollection enables tool-call collection in a context.
func WithToolCallsCollection(ctx context.Context) context.Context {
	toolCalls := make([]ToolCall, 0)
	return context.WithValue(ctx, toolCallsKey{}, &toolCalls)
}

// AddToolCallToContext appends a tool call to the context collection.
func AddToolCallToContext(ctx context.Context, toolCall ToolCall) {
	if toolCalls, ok := ctx.Value(toolCallsKey{}).(*[]ToolCall); ok {
		*toolCalls = append(*toolCalls, toolCall)
	}
}

// GetToolCallsFromContext returns the collected tool calls.
func GetToolCallsFromContext(ctx context.Context) []ToolCall {
	if toolCalls, ok := ctx.Value(toolCallsKey{}).(*[]ToolCall); ok {
		return *toolCalls
	}
	return nil
}
