package agent

import (
	"context"
	"time"

	"nu/internal/contracts"
)

// AgentContextKey represents a key for values in AgentContext.
type AgentContextKey string

const (
	// OrganizationIDKey is the key for the organization ID
	AgentContextOrganizationIDKey AgentContextKey = "organization_id"

	// ConversationIDKey is the key for the conversation ID
	AgentContextConversationIDKey AgentContextKey = "conversation_id"

	// UserIDKey is the key for the user ID
	AgentContextUserIDKey AgentContextKey = "user_id"

	// RequestIDKey is the key for the request ID
	AgentContextRequestIDKey AgentContextKey = "request_id"

	// MemoryKey is the key for the memory
	AgentContextMemoryKey AgentContextKey = "memory"

	// ToolsKey is the key for the tools
	AgentContextToolsKey AgentContextKey = "tools"

	// DataStoreKey is the key for the data store
	AgentContextDataStoreKey AgentContextKey = "data_store"

	// VectorStoreKey is the key for the vector store
	AgentContextVectorStoreKey AgentContextKey = "vector_store"

	// LLMKey is the key for the LLM
	AgentContextLLMKey AgentContextKey = "llm"

	// EnvironmentKey is the key for environment variables
	AgentContextEnvironmentKey AgentContextKey = "environment"
)

// AgentContext represents the context for an agent
type AgentContext struct {
	ctx context.Context
}

// NewAgentContext creates an agent context with a background parent.
func NewAgentContext() *AgentContext {
	return &AgentContext{
		ctx: context.Background(),
	}
}

// AgentContextFromContext wraps a standard context.
func AgentContextFromContext(ctx context.Context) *AgentContext {
	return &AgentContext{
		ctx: ctx,
	}
}

// WithOrganizationID sets the organization ID in the context
func (c *AgentContext) WithOrganizationID(orgID string) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextOrganizationIDKey, orgID)
	return c
}

// OrganizationID returns the organization ID from the context
func (c *AgentContext) OrganizationID() (string, bool) {
	orgID, ok := c.ctx.Value(AgentContextOrganizationIDKey).(string)
	return orgID, ok
}

// WithConversationID sets the conversation ID in the context
func (c *AgentContext) WithConversationID(conversationID string) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextConversationIDKey, conversationID)
	return c
}

// ConversationID returns the conversation ID from the context
func (c *AgentContext) ConversationID() (string, bool) {
	conversationID, ok := c.ctx.Value(AgentContextConversationIDKey).(string)
	return conversationID, ok
}

// WithUserID sets the user ID in the context
func (c *AgentContext) WithUserID(userID string) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextUserIDKey, userID)
	return c
}

// UserID returns the user ID from the context
func (c *AgentContext) UserID() (string, bool) {
	userID, ok := c.ctx.Value(AgentContextUserIDKey).(string)
	return userID, ok
}

// WithRequestID sets the request ID in the context
func (c *AgentContext) WithRequestID(requestID string) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextRequestIDKey, requestID)
	return c
}

// RequestID returns the request ID from the context
func (c *AgentContext) RequestID() (string, bool) {
	requestID, ok := c.ctx.Value(AgentContextRequestIDKey).(string)
	return requestID, ok
}

// WithMemory sets the memory in the context
func (c *AgentContext) WithMemory(memory contracts.Memory) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextMemoryKey, memory)
	return c
}

// Memory returns the memory from the context
func (c *AgentContext) Memory() (contracts.Memory, bool) {
	memory, ok := c.ctx.Value(AgentContextMemoryKey).(contracts.Memory)
	return memory, ok
}

// WithTools sets the tools in the context
func (c *AgentContext) WithTools(tools contracts.ToolRegistry) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextToolsKey, tools)
	return c
}

// Tools returns the tools from the context
func (c *AgentContext) Tools() (contracts.ToolRegistry, bool) {
	tools, ok := c.ctx.Value(AgentContextToolsKey).(contracts.ToolRegistry)
	return tools, ok
}

// WithDataStore sets the data store in the context
func (c *AgentContext) WithDataStore(dataStore contracts.DataStore) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextDataStoreKey, dataStore)
	return c
}

// DataStore returns the data store from the context
func (c *AgentContext) DataStore() (contracts.DataStore, bool) {
	dataStore, ok := c.ctx.Value(AgentContextDataStoreKey).(contracts.DataStore)
	return dataStore, ok
}

// WithVectorStore sets the vector store in the context
func (c *AgentContext) WithVectorStore(vectorStore contracts.VectorStore) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextVectorStoreKey, vectorStore)
	return c
}

// VectorStore returns the vector store from the context
func (c *AgentContext) VectorStore() (contracts.VectorStore, bool) {
	vectorStore, ok := c.ctx.Value(AgentContextVectorStoreKey).(contracts.VectorStore)
	return vectorStore, ok
}

// WithLLM sets the LLM in the context
func (c *AgentContext) WithLLM(llm contracts.LLM) *AgentContext {
	c.ctx = context.WithValue(c.ctx, AgentContextLLMKey, llm)
	return c
}

// LLM returns the LLM from the context
func (c *AgentContext) LLM() (contracts.LLM, bool) {
	llm, ok := c.ctx.Value(AgentContextLLMKey).(contracts.LLM)
	return llm, ok
}

// WithEnvironment sets an environment variable in the context
func (c *AgentContext) WithEnvironment(key string, value interface{}) *AgentContext {
	env, ok := c.ctx.Value(AgentContextEnvironmentKey).(map[string]interface{})
	if !ok {
		env = make(map[string]interface{})
	}
	env[key] = value
	c.ctx = context.WithValue(c.ctx, AgentContextEnvironmentKey, env)
	return c
}

// Environment returns an environment variable from the context
func (c *AgentContext) Environment(key string) (interface{}, bool) {
	env, ok := c.ctx.Value(AgentContextEnvironmentKey).(map[string]interface{})
	if !ok {
		return nil, false
	}
	value, ok := env[key]
	return value, ok
}

// WithTimeout sets a timeout for the context
func (c *AgentContext) WithTimeout(timeout time.Duration) (*AgentContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.ctx, timeout) // #nosec G118 - cancel func is returned to caller
	return &AgentContext{ctx: ctx}, cancel
}

// WithDeadline sets a deadline for the context
func (c *AgentContext) WithDeadline(deadline time.Time) (*AgentContext, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.ctx, deadline) // #nosec G118 - cancel func is returned to caller
	return &AgentContext{ctx: ctx}, cancel
}

// WithCancel returns a new context that can be canceled
func (c *AgentContext) WithCancel() (*AgentContext, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.ctx) // #nosec G118 - cancel func is returned to caller
	return &AgentContext{ctx: ctx}, cancel
}

// Context returns the underlying context.Context
func (c *AgentContext) Context() context.Context {
	return c.ctx
}

// Done returns the done channel from the context
func (c *AgentContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err returns the error from the context
func (c *AgentContext) Err() error {
	return c.ctx.Err()
}

// Deadline returns the deadline from the context
func (c *AgentContext) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

// Value returns a value from the context
func (c *AgentContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}
