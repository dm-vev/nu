package mcp

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/tool"
)

// CollectTools lists tools exposed by connected MCP servers.
func (m *Manager) CollectTools(ctx context.Context) ([]contracts.Tool, error) {
	var mcpTools []contracts.Tool
	for _, server := range m.servers {
		tools, err := server.ListTools(ctx)
		if err != nil {
			m.logger.Error(ctx, fmt.Sprintf("Failed to list tools from MCP server: %v", err), nil)
			continue
		}
		for _, mcpTool := range tools {
			mcpTools = append(mcpTools, tool.NewTool(mcpTool.Name, mcpTool.Description, mcpTool.Schema, server))
		}
	}
	return mcpTools, nil
}
