package mcp

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

// Initialize eagerly discovers regular and lazy MCP tools.
func (m *Manager) Initialize(ctx context.Context) ([]contracts.Tool, error) {
	var allTools []contracts.Tool
	if len(m.servers) > 0 {
		mcpTools, err := m.CollectTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to collect MCP server tools: %w", err)
		}
		allTools = append(allTools, mcpTools...)
		m.logger.Info(context.Background(), fmt.Sprintf("Initialized %d MCP server tools", len(mcpTools)), nil)
	}
	if len(m.lazyConfigs) > 0 {
		lazyTools := m.LazyTools()
		allTools = append(allTools, lazyTools...)
		m.logger.Info(context.Background(), fmt.Sprintf("Initialized %d lazy MCP tools", len(lazyTools)), nil)
	}
	return deduplicateTools(allTools), nil
}

func deduplicateTools(tools []contracts.Tool) []contracts.Tool {
	seen := make(map[string]bool, len(tools))
	result := make([]contracts.Tool, 0, len(tools))
	for _, candidate := range tools {
		if candidate == nil || candidate.Name() == "" || seen[candidate.Name()] {
			continue
		}
		seen[candidate.Name()] = true
		result = append(result, candidate)
	}
	return result
}
