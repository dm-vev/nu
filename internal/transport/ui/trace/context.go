package trace

import (
	"context"

	memory "nu/internal/memory/conversation"
	"nu/internal/multitenancy"
)

func (c *Collector) getParentSpanID(ctx context.Context) string {
	if spanID, ok := ctx.Value(spanIDKey{}).(string); ok {
		return spanID
	}
	return ""
}

func (c *Collector) getConversationID(ctx context.Context) string {
	if id, ok := memory.GetConversationID(ctx); ok {
		return id
	}
	return ""
}

func (c *Collector) getOrgID(ctx context.Context) string {
	if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
		return orgID
	}
	return ""
}

// Context key for span ID
type spanIDKey struct{}

// noOpSpan is a no-op implementation of contracts.Span
type noOpSpan struct{}

func (s *noOpSpan) End()                                                    {}
func (s *noOpSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *noOpSpan) SetAttribute(key string, value interface{})              {}
func (s *noOpSpan) RecordError(err error)                                   {}

// Utility functions
func uiContains(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) && s[:len(substr)] == substr {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
