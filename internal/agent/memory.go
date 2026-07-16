package agent

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	memoryfactory "nu/internal/memory/factory"
)

// GetMemory returns the memory instance (for use in custom functions)
func (a *Agent) GetMemory() contracts.Memory {
	return a.memory
}

// GetAllConversations returns all conversation IDs from memory
func (a *Agent) GetAllConversations(ctx context.Context) ([]string, error) {
	if a.memory == nil {
		return []string{}, nil
	}

	// Check if memory supports conversation operations
	if convMem, ok := a.memory.(contracts.ConversationMemory); ok {
		return convMem.GetAllConversations(ctx)
	}

	// Fallback: return empty list for memories that don't support conversations
	return []string{}, nil
}

// GetConversationMessages gets all messages for a specific conversation
func (a *Agent) GetConversationMessages(ctx context.Context, conversationID string) ([]contracts.Message, error) {
	if a.memory == nil {
		return []contracts.Message{}, nil
	}

	// Check if memory supports conversation operations
	if convMem, ok := a.memory.(contracts.ConversationMemory); ok {
		return convMem.GetConversationMessages(ctx, conversationID)
	}

	// Fallback: return empty list for memories that don't support conversations
	return []contracts.Message{}, nil
}

// GetMemoryStatistics returns basic memory statistics
func (a *Agent) GetMemoryStatistics(ctx context.Context) (totalConversations, totalMessages int, err error) {
	if a.memory == nil {
		return 0, 0, nil
	}

	// Check if memory supports conversation operations
	if convMem, ok := a.memory.(contracts.ConversationMemory); ok {
		return convMem.GetMemoryStatistics(ctx)
	}

	// Fallback: return basic stats for memories that don't support conversations
	return 0, 0, nil
}

// CreateMemoryFromConfig creates a memory instance from YAML configuration
// This function is intended to be used by agent-blueprint applications that need
// to instantiate memory from YAML config stored in the agent
func CreateMemoryFromConfig(memoryConfig map[string]interface{}, llmClient contracts.LLM) (contracts.Memory, error) {
	if memoryConfig == nil {
		return nil, fmt.Errorf("memory config is nil")
	}

	factory := memoryfactory.NewFactory()
	return factory.CreateMemory(memoryConfig, llmClient)
}

// GetMemoryConfig returns the stored memory configuration from YAML
// This allows agent-blueprint to access the memory config for instantiation
func (a *Agent) GetMemoryConfig() map[string]interface{} {
	return a.memoryConfig
}
