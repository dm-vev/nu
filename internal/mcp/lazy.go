package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// LazyMCPTool is a tool that initializes its MCP server on first use
type LazyMCPTool struct {
	name         string
	description  string
	schema       interface{}
	schemaLoaded bool
	serverConfig LazyMCPServerConfig
	serverInfo   *contracts.MCPServerInfo
	logger       telemetry.Logger
	mu           sync.RWMutex
}

// NewLazyMCPTool creates a new lazy MCP tool
func NewLazyMCPTool(name, description string, schema interface{}, config LazyMCPServerConfig) contracts.Tool {
	tool := &LazyMCPTool{
		name: name, description: description, schema: nil,
		schemaLoaded: false, serverConfig: config,
	}
	if config.Logger != nil {
		tool.logger = config.Logger
	} else {
		tool.logger = telemetry.NewLogger()
	}
	return tool
}

func (t *LazyMCPTool) Name() string        { return t.name }
func (t *LazyMCPTool) DisplayName() string { return t.name }

func (t *LazyMCPTool) Description() string {
	if t.description != "" {
		return t.description
	}
	if t.serverInfo != nil && t.serverInfo.Title != "" {
		return fmt.Sprintf("%s (from %s)", t.name, t.serverInfo.Title)
	}
	return fmt.Sprintf("%s (MCP tool)", t.name)
}

func (t *LazyMCPTool) Internal() bool { return false }

func (t *LazyMCPTool) getServer(ctx context.Context) (contracts.MCPServer, error) {
	server, err := globalServerCache.getOrCreateServer(ctx, t.serverConfig)
	if err != nil {
		return nil, err
	}
	if t.serverInfo == nil {
		serverKey := fmt.Sprintf("%s:%s:%v", t.serverConfig.Type, t.serverConfig.Name, t.serverConfig.Command)
		globalServerCache.mu.RLock()
		if metadata, exists := globalServerCache.serverMetadata[serverKey]; exists {
			t.serverInfo = metadata
		}
		globalServerCache.mu.RUnlock()
	}
	return server, nil
}

// Run executes the tool with the given input
func (t *LazyMCPTool) Run(ctx context.Context, input string) (string, error) {
	server, err := t.getServer(ctx)
	if err != nil {
		return "", err
	}
	var args map[string]interface{}
	if input != "" {
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return "", fmt.Errorf("failed to parse input as JSON: %w", err)
		}
	}
	t.logger.Info(ctx, "[MCP TOOL CALL] Calling MCP tool", map[string]interface{}{
		"tool_name": t.name, "args": args, "server_name": t.serverConfig.Name,
		"server_type": t.serverConfig.Type, "command": t.serverConfig.Command,
		"env_count": len(t.serverConfig.Env),
	})
	for _, envVar := range t.serverConfig.Env {
		if len(envVar) > 50 {
			t.logger.Debug(ctx, "[MCP ENV] Server environment variable", map[string]interface{}{
				"env_var": envVar[:50] + "...",
			})
		} else {
			t.logger.Debug(ctx, "[MCP ENV] Server environment variable", map[string]interface{}{"env_var": envVar})
		}
	}

	resp, err := server.CallTool(ctx, t.name, args)
	if err != nil {
		t.logger.Error(ctx, "[MCP TOOL ERROR] MCP tool call failed with error", map[string]interface{}{
			"tool_name": t.name, "server_name": t.serverConfig.Name,
			"error": err.Error(), "error_type": fmt.Sprintf("%T", err),
		})
		return "", fmt.Errorf("MCP server call failed: %v", err)
	}
	t.logger.Info(ctx, "[MCP TOOL RESPONSE] Received MCP tool response", map[string]interface{}{
		"tool_name": t.name, "server_name": t.serverConfig.Name,
		"is_error": resp.IsError, "content": resp.Content,
		"content_type": fmt.Sprintf("%T", resp.Content),
	})
	if resp.IsError {
		var errorMsg string
		switch content := resp.Content.(type) {
		case string:
			errorMsg = content
			t.logger.Error(ctx, "[MCP TOOL ERROR] MCP server returned error (string)", map[string]interface{}{
				"tool_name": t.name, "server_name": t.serverConfig.Name, "error": content,
			})
		case []byte:
			errorMsg = string(content)
			t.logger.Error(ctx, "[MCP TOOL ERROR] MCP server returned error (bytes)", map[string]interface{}{
				"tool_name": t.name, "server_name": t.serverConfig.Name, "error": errorMsg,
			})
		case map[string]interface{}:
			if msg, ok := content["message"].(string); ok {
				errorMsg = msg
			} else if bytes, err := json.Marshal(content); err == nil {
				errorMsg = string(bytes)
			} else {
				errorMsg = fmt.Sprintf("%v", content)
			}
			t.logger.Error(ctx, "[MCP TOOL ERROR] MCP server returned error (map)", map[string]interface{}{
				"tool_name": t.name, "server_name": t.serverConfig.Name,
				"error_content": content, "parsed_message": errorMsg,
			})
		case []interface{}:
			if bytes, err := json.Marshal(content); err == nil {
				errorMsg = string(bytes)
			} else {
				errorMsg = fmt.Sprintf("%v", content)
			}
			t.logger.Error(ctx, "[MCP TOOL ERROR] MCP server returned error (array)", map[string]interface{}{
				"tool_name": t.name, "server_name": t.serverConfig.Name,
				"error_content": content, "parsed_error": errorMsg, "array_length": len(content),
			})
		default:
			if bytes, err := json.Marshal(content); err == nil {
				errorMsg = string(bytes)
			} else {
				errorMsg = fmt.Sprintf("%v", content)
			}
			t.logger.Error(ctx, "[MCP TOOL ERROR] MCP server returned error (unknown type)", map[string]interface{}{
				"tool_name": t.name, "server_name": t.serverConfig.Name,
				"error_type": fmt.Sprintf("%T", content), "error": errorMsg, "raw_content": content,
			})
		}
		return "", fmt.Errorf("MCP tool error from server '%s': %s", t.serverConfig.Name, errorMsg)
	}
	return extractTextFromMCPContent(resp.Content), nil
}

func (t *LazyMCPTool) Execute(ctx context.Context, args string) (string, error) {
	return t.Run(ctx, args)
}
