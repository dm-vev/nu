package agent

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/tools/graphrag"
)

// WithGraphRAG adds GraphRAG capabilities to the agent.
// When a GraphRAGStore is provided, the agent automatically registers
// GraphRAG tools (search, add_entity, add_relationship, get_context, extract).
func WithGraphRAG(store contracts.GraphRAGStore) Option {
	return func(a *Agent) {
		if store == nil {
			return
		}

		// Store reference for direct access
		a.graphRAGStore = store

		// Create and register GraphRAG tools
		configuredTools := createGraphRAGTools(store, a.llm)
		a.tools = deduplicateTools(append(a.tools, configuredTools...))

		if a.logger != nil {
			a.logger.Info(context.Background(), "GraphRAG enabled", map[string]interface{}{
				"tools": len(configuredTools),
			})
		}
	}
}

func createGraphRAGTools(store contracts.GraphRAGStore, llm contracts.LLM) []contracts.Tool {
	tools := []contracts.Tool{
		graphrag.NewTool(store),
		graphrag.NewAddEntityTool(store),
		graphrag.NewAddRelationshipTool(store),
		graphrag.NewGetContextTool(store),
	}
	if llm != nil {
		tools = append(tools, graphrag.NewExtractTool(store, llm))
	}
	return tools
}

// GetGraphRAGStore returns the GraphRAG store if configured.
// Returns nil if GraphRAG is not enabled.
func (a *Agent) GetGraphRAGStore() contracts.GraphRAGStore {
	return a.graphRAGStore
}

// HasGraphRAG returns true if the agent has GraphRAG capabilities enabled.
func (a *Agent) HasGraphRAG() bool {
	return a.graphRAGStore != nil
}
