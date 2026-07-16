package agent

import (
	"context"

	"nu/internal/contracts"
	"nu/internal/tools/graphrag"
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
		graphTools := createGraphRAGTools(store, a.llm)
		a.tools = deduplicateTools(append(a.tools, graphTools...))

		if a.logger != nil {
			a.logger.Info(context.Background(), "GraphRAG enabled", map[string]interface{}{
				"tools": len(graphTools),
			})
		}
	}
}

// createGraphRAGTools creates the standard GraphRAG tools.
func createGraphRAGTools(store contracts.GraphRAGStore, llm contracts.LLM) []contracts.Tool {
	graphTools := []contracts.Tool{
		graphrag.NewTool(store),
		graphrag.NewAddEntityTool(store),
		graphrag.NewAddRelationshipTool(store),
		graphrag.NewGetContextTool(store),
	}

	// Only add extract tool if LLM is available
	if llm != nil {
		graphTools = append(graphTools, graphrag.NewExtractTool(store, llm))
	}

	return graphTools
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
