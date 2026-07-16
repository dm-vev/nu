package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/schema"
	"github.com/dm-vev/nu/telemetry"
)

// ToolManager provides enhanced tool operations with schema support
type ToolManager struct {
	servers   []contracts.MCPServer
	validator *schema.SchemaValidator
	logger    telemetry.Logger
}

func NewToolManager(servers []contracts.MCPServer) *ToolManager {
	return &ToolManager{servers: servers, validator: schema.NewSchemaValidator(), logger: telemetry.NewLogger()}
}

func (tm *ToolManager) ListAllTools(ctx context.Context) (map[string][]contracts.MCPTool, error) {
	result := make(map[string][]contracts.MCPTool)
	for i, server := range tm.servers {
		serverName := fmt.Sprintf("server-%d", i)
		tools, err := server.ListTools(ctx)
		if err != nil {
			tm.logger.Warn(ctx, "Failed to list tools from server", map[string]interface{}{
				"server": serverName, "error": err.Error(),
			})
			continue
		}
		result[serverName] = tools
		tm.logger.Debug(ctx, "Listed tools from server", map[string]interface{}{
			"server": serverName, "tool_count": len(tools),
		})
	}
	return result, nil
}

func (tm *ToolManager) CallToolWithValidation(ctx context.Context, toolName string, args interface{}) (*contracts.MCPToolResponse, error) {
	var tool *contracts.MCPTool
	var server contracts.MCPServer
	for _, srv := range tm.servers {
		tools, err := srv.ListTools(ctx)
		if err != nil {
			continue
		}
		for _, candidate := range tools {
			if candidate.Name == toolName {
				tool = &candidate
				server = srv
				break
			}
		}
		if tool != nil {
			break
		}
	}
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	response, err := server.CallTool(ctx, toolName, args)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}
	if err := tm.validator.ValidateToolResponse(ctx, *tool, response); err != nil {
		tm.logger.Error(ctx, "Tool response validation failed", map[string]interface{}{
			"tool_name": toolName, "error": err.Error(),
		})
	}
	tm.logger.Debug(ctx, "Tool called successfully", map[string]interface{}{
		"tool_name": toolName, "has_structured": response.StructuredContent != nil,
		"has_output_schema": tool.OutputSchema != nil,
	})
	return response, nil
}

func (tm *ToolManager) GetToolsByCategory(ctx context.Context, category string) ([]ToolMatch, error) {
	var matches []ToolMatch
	for _, server := range tm.servers {
		tools, err := server.ListTools(ctx)
		if err != nil {
			continue
		}
		for _, tool := range tools {
			if tm.matchesCategory(tool, category) {
				matches = append(matches, ToolMatch{Server: server, Tool: tool})
			}
		}
	}
	return matches, nil
}

func (tm *ToolManager) GetToolsWithOutputSchema(ctx context.Context) ([]ToolMatch, error) {
	var matches []ToolMatch
	for i, server := range tm.servers {
		serverName := fmt.Sprintf("server-%d", i)
		tools, err := server.ListTools(ctx)
		if err != nil {
			tm.logger.Warn(ctx, "Failed to list tools from server", map[string]interface{}{
				"server": serverName, "error": err.Error(),
			})
			continue
		}
		for _, tool := range tools {
			if tool.OutputSchema != nil {
				matches = append(matches, ToolMatch{Server: server, Tool: tool})
			}
		}
	}
	tm.logger.Debug(ctx, "Found tools with output schemas", map[string]interface{}{"count": len(matches)})
	return matches, nil
}

type ToolMatch struct {
	Server contracts.MCPServer
	Tool   contracts.MCPTool
}

func (tm *ToolManager) matchesCategory(tool contracts.MCPTool, category string) bool {
	return containsIgnoreCaseFunc(fmt.Sprintf("%s %s", tool.Name, tool.Description), category)
}

func containsIgnoreCaseFunc(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
