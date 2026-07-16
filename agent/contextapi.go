package agent

import (
	"context"
	"time"

	agentcontext "github.com/dm-vev/nu/agent/context"
)

type AgentContext = agentcontext.AgentContext
type AgentContextKey = agentcontext.AgentContextKey
type ContextKey = agentcontext.ContextKey
type SubAgentContext = agentcontext.SubAgentContext

const (
	SubAgentNameKey        = agentcontext.SubAgentNameKey
	ParentAgentKey         = agentcontext.ParentAgentKey
	RecursionDepthKey      = agentcontext.RecursionDepthKey
	InvocationIDKey        = agentcontext.InvocationIDKey
	MaxRecursionDepth      = agentcontext.MaxRecursionDepth
	DefaultSubAgentTimeout = agentcontext.DefaultSubAgentTimeout

	AgentContextOrganizationIDKey = agentcontext.AgentContextOrganizationIDKey
	AgentContextConversationIDKey = agentcontext.AgentContextConversationIDKey
	AgentContextUserIDKey         = agentcontext.AgentContextUserIDKey
	AgentContextRequestIDKey      = agentcontext.AgentContextRequestIDKey
	AgentContextMemoryKey         = agentcontext.AgentContextMemoryKey
	AgentContextToolsKey          = agentcontext.AgentContextToolsKey
	AgentContextDataStoreKey      = agentcontext.AgentContextDataStoreKey
	AgentContextVectorStoreKey    = agentcontext.AgentContextVectorStoreKey
	AgentContextLLMKey            = agentcontext.AgentContextLLMKey
	AgentContextEnvironmentKey    = agentcontext.AgentContextEnvironmentKey
)

func NewAgentContext() *AgentContext { return agentcontext.NewAgentContext() }

func AgentContextFromContext(ctx context.Context) *AgentContext {
	return agentcontext.AgentContextFromContext(ctx)
}

func WithSubAgentContext(ctx context.Context, parentAgent, subAgentName string) context.Context {
	return agentcontext.WithSubAgentContext(ctx, parentAgent, subAgentName)
}

func GetRecursionDepth(ctx context.Context) int  { return agentcontext.GetRecursionDepth(ctx) }
func GetSubAgentName(ctx context.Context) string { return agentcontext.GetSubAgentName(ctx) }
func GetParentAgent(ctx context.Context) string  { return agentcontext.GetParentAgent(ctx) }
func GetInvocationID(ctx context.Context) string { return agentcontext.GetInvocationID(ctx) }
func IsSubAgentCall(ctx context.Context) bool    { return agentcontext.IsSubAgentCall(ctx) }
func ValidateRecursionDepth(ctx context.Context) error {
	return agentcontext.ValidateRecursionDepth(ctx)
}
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return agentcontext.WithTimeout(ctx, timeout)
}
