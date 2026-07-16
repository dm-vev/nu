package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nu/internal/contracts"
)

func (t *LazyMCPTool) discoverSchema(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.schemaLoaded {
		return nil
	}
	server, err := t.getServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MCP server: %w", err)
	}
	tools, err := server.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools from MCP server: %w", err)
	}
	for _, tool := range tools {
		if tool.Name == t.name {
			t.schema = tool.Schema
			t.schemaLoaded = true
			t.logger.Debug(ctx, "Discovered schema for MCP tool", map[string]interface{}{
				"tool_name": t.name, "schema": tool.Schema,
			})
			return nil
		}
	}
	t.logger.Warn(ctx, "Tool not found in MCP server tool list", map[string]interface{}{"tool_name": t.name})
	t.schemaLoaded = true
	return nil
}

// Parameters returns the parameters that the tool accepts
func (t *LazyMCPTool) Parameters() map[string]contracts.ParameterSpec {
	ctx := context.Background()
	if !t.schemaLoaded {
		if err := t.discoverSchema(ctx); err != nil {
			t.logger.Warn(ctx, "Failed to discover schema for tool", map[string]interface{}{
				"tool_name": t.name, "error": err.Error(),
			})
			return make(map[string]contracts.ParameterSpec)
		}
	}
	params := make(map[string]contracts.ParameterSpec)
	var schemaMap map[string]interface{}
	switch schema := t.schema.(type) {
	case map[string]interface{}:
		schemaMap = schema
	case string:
		if err := json.Unmarshal([]byte(schema), &schemaMap); err != nil {
			t.logger.Warn(ctx, "Failed to parse schema JSON string", map[string]interface{}{
				"tool_name": t.name, "error": err.Error(),
			})
			return params
		}
	default:
		if schemaBytes, err := json.Marshal(t.schema); err == nil {
			if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
				t.logger.Warn(ctx, "Failed to unmarshal schema after marshaling", map[string]interface{}{
					"tool_name": t.name, "error": err.Error(),
				})
				return params
			}
		} else {
			t.logger.Warn(ctx, "Schema cannot be marshaled to JSON", map[string]interface{}{
				"tool_name": t.name, "schema_type": fmt.Sprintf("%T", t.schema),
			})
			return params
		}
	}

	if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for name, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				var paramType any
				if typeVal, ok := propMap["type"]; ok && typeVal != nil {
					paramType = typeVal
				} else if anyOf, ok := propMap["anyOf"].([]interface{}); ok && len(anyOf) > 0 {
					for _, typeOption := range anyOf {
						if typeMap, ok := typeOption.(map[string]interface{}); ok {
							if value, ok := typeMap["type"].(string); ok && value != "null" {
								paramType = value
								break
							}
						}
					}
					if paramType == nil {
						paramType = "string"
					}
				} else {
					paramType = "string"
				}

				paramSpec := contracts.ParameterSpec{
					Type: paramType, Description: fmt.Sprintf("%v", propMap["description"]),
				}
				if paramType == "array" || strings.Contains(fmt.Sprintf("%v", paramType), "array") {
					paramSpec.Items = &contracts.ParameterSpec{Type: "string"}
					if itemsMap, ok := propMap["items"].(map[string]interface{}); ok {
						if itemType, ok := itemsMap["type"].(string); ok && itemType != "" {
							paramSpec.Items.Type = itemType
						}
						if enum, ok := itemsMap["enum"].([]interface{}); ok {
							paramSpec.Items.Enum = enum
						}
					}
				}
				if enum, ok := propMap["enum"].([]interface{}); ok {
					paramSpec.Enum = enum
				}
				if defaultVal, ok := propMap["default"]; ok {
					paramSpec.Default = defaultVal
				}
				if required, ok := schemaMap["required"].([]interface{}); ok {
					for _, req := range required {
						if req == name {
							paramSpec.Required = true
							break
						}
					}
				}
				params[name] = paramSpec
			}
		}
	}
	return params
}
