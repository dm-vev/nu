package tools

import (
	"context"

	"nu/internal/contracts"
)

// Context keys for sub-agent metadata
type contextKey string

const (
	recursionDepthKey contextKey = "recursion_depth"
	subAgentNameKey   contextKey = "sub_agent_name"
	parentAgentKey    contextKey = "parent_agent"
	invocationIDKey   contextKey = "invocation_id"

	// MaxRecursionDepth is the maximum allowed recursion depth
	MaxRecursionDepth = 5
)

// getRecursionDepth retrieves the current recursion depth from context
func getRecursionDepth(ctx context.Context) int {
	if depth, ok := ctx.Value(recursionDepthKey).(int); ok {
		return depth
	}
	return 0
}

// withSubAgentContext adds sub-agent context information for testing purposes
func withSubAgentContext(ctx context.Context, parentAgent, subAgentName string) context.Context {
	depth := getRecursionDepth(ctx)
	ctx = context.WithValue(ctx, subAgentNameKey, subAgentName)
	ctx = context.WithValue(ctx, parentAgentKey, parentAgent)
	ctx = context.WithValue(ctx, recursionDepthKey, depth+1)
	return ctx
}

// WithStreamForwarder adds a stream forwarder to the context
func WithStreamForwarder(ctx context.Context, forwarder contracts.StreamForwarder) context.Context {
	return context.WithValue(ctx, contracts.StreamForwarderKey, forwarder)
}
