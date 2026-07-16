package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nu/internal/contracts"
)

// ListTools lists the tools available on the MCP server
func (s *Server) ListTools(ctx context.Context) ([]contracts.MCPTool, error) {
	s.logger.Debug(ctx, "Listing MCP tools", nil)
	resp, err := s.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		mcpErr := ClassifyError(err, "ListTools", "server", "unknown")
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(ctx, "Failed to list MCP tools", map[string]interface{}{
			"error":      err.Error(),
			"error_type": mcpErr.ErrorType,
			"retryable":  mcpErr.Retryable,
		})
		return nil, mcpErr
	}

	tools := make([]contracts.MCPTool, 0, len(resp.Tools))
	for _, t := range resp.Tools {
		tool := contracts.MCPTool{
			Name:        t.Name,
			Description: t.Description,
			Schema:      t.InputSchema,
			Metadata:    make(map[string]interface{}),
		}
		if t.Annotations != nil {
			tool.Metadata["annotations"] = "present"
		}
		tools = append(tools, tool)
	}
	s.logger.Info(ctx, "Successfully listed MCP tools", map[string]interface{}{
		"tool_count": len(tools),
	})
	return tools, nil
}

// CallTool calls a tool on the MCP server
func (s *Server) CallTool(ctx context.Context, name string, args interface{}) (*contracts.MCPToolResponse, error) {
	s.logger.Debug(ctx, "Calling MCP tool", map[string]interface{}{
		"tool_name": name,
		"args":      args,
	})
	params := &mcp.CallToolParams{Name: name, Arguments: args}
	s.logger.Debug(ctx, "Calling session.CallTool", map[string]interface{}{
		"tool_name": name,
		"params":    params,
	})

	resp, err := s.session.CallTool(ctx, params)
	if err != nil {
		mcpErr := ClassifyError(err, "CallTool", "server", "unknown")
		_ = mcpErr.WithMetadata("tool_name", name)
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(ctx, "Failed to call MCP tool", map[string]interface{}{
			"tool_name":  name,
			"error":      err.Error(),
			"error_type": mcpErr.ErrorType,
			"retryable":  mcpErr.Retryable,
		})
		return nil, mcpErr
	}

	s.logger.Info(ctx, "[MCP SERVER] Received response from session.CallTool", map[string]interface{}{
		"tool_name":    name,
		"is_error":     resp.IsError,
		"content":      resp.Content,
		"content_type": fmt.Sprintf("%T", resp.Content),
		"meta":         resp.Meta,
	})
	if resp.IsError {
		contentJSON, _ := json.Marshal(resp.Content)
		s.logger.Error(ctx, "[MCP SERVER ERROR] MCP tool returned error", map[string]interface{}{
			"tool_name":    name,
			"content":      resp.Content,
			"content_type": fmt.Sprintf("%T", resp.Content),
			"content_json": string(contentJSON),
			"is_error":     resp.IsError,
			"meta":         resp.Meta,
		})
	} else {
		s.logger.Info(ctx, "[MCP SERVER SUCCESS] MCP tool executed successfully", map[string]interface{}{
			"tool_name": name,
		})
	}

	response := &contracts.MCPToolResponse{
		Content:  resp.Content,
		IsError:  resp.IsError,
		Metadata: make(map[string]interface{}),
	}
	if resp.Meta != nil {
		response.Metadata = resp.Meta
	}
	if metadata, ok := resp.Meta["structuredContent"]; ok {
		response.StructuredContent = metadata
	}
	return response, nil
}
