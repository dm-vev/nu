package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"nu/internal/contracts"

	"github.com/google/jsonschema-go/jsonschema"
)

// Tool implements contracts.Tool for MCP tools
type Tool struct {
	name        string
	description string
	schema      interface{}
	server      contracts.MCPServer
}

// NewTool creates a new Tool
func NewTool(name, description string, schema interface{}, server contracts.MCPServer) contracts.Tool {
	return &Tool{
		name:        name,
		description: description,
		schema:      schema,
		server:      server,
	}
}

// Name returns the name of the tool
func (t *Tool) Name() string {
	return t.name
}

// DisplayName implements contracts.ToolWithDisplayName.DisplayName
func (t *Tool) DisplayName() string {
	// MCP tools can use the name as display name
	return t.name
}

// Description returns a description of what the tool does
func (t *Tool) Description() string {
	return t.description
}

// Internal implements contracts.InternalTool.Internal
func (t *Tool) Internal() bool {
	// MCP tools are typically visible to users
	return false
}

// Run executes the tool with the given input
func (t *Tool) Run(ctx context.Context, input string) (string, error) {
	// Parse the input as JSON to get the arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("failed to parse input as JSON: %w", err)
	}

	// Call the tool on the MCP server
	resp, err := t.server.CallTool(ctx, t.name, args)
	if err != nil {
		return "", err
	}

	// Convert the response to a string
	if resp.IsError {
		return "", fmt.Errorf("MCP tool error: %v", resp.Content)
	}

	// Try to convert the content to a string
	switch content := resp.Content.(type) {
	case string:
		return content, nil
	case []byte:
		return string(content), nil
	default:
		// Try to JSON marshal the content
		bytes, err := json.Marshal(content)
		if err != nil {
			return fmt.Sprintf("%v", content), nil
		}
		return string(bytes), nil
	}
}

// Parameters returns the parameters that the tool accepts
func (t *Tool) Parameters() map[string]contracts.ParameterSpec {
	// Convert the schema to a map of ParameterSpec
	// This is a simplified implementation; in a real implementation,
	// we would parse the JSON schema to extract parameter information
	params := make(map[string]contracts.ParameterSpec)
	switch toolSchema := t.schema.(type) {
	// For backward compatibility
	//TODO remove in future releases
	case map[string]interface{}:
		// If the schema is a string, try to parse it as JSON
		// Try to convert the schema to a map
		if properties, ok := toolSchema["properties"].(map[string]interface{}); ok {
			for name, prop := range properties {
				if propMap, ok := prop.(map[string]interface{}); ok {
					paramSpec := contracts.ParameterSpec{
						Type:        propMap["type"],
						Description: fmt.Sprintf("%v", propMap["description"]),
					}

					// Array parameters must carry an items specification;
					// downstream converters (Gemini, OpenAI) reject
					// function declarations that expose an `array` without
					// `items`. Default to string items when the server's
					// schema omits or malforms it.
					if typeStr := asString(propMap["type"]); typeStr == "array" {
						paramSpec.Items = extractItemsFromMap(propMap["items"])
					}

					if enum, ok := propMap["enum"].([]interface{}); ok {
						paramSpec.Enum = enum
					}

					// Check if the parameter is required
					if required, ok := toolSchema["required"].([]interface{}); ok {
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
	case *jsonschema.Schema:
		for name, prop := range toolSchema.Properties {
			paramSpec := contracts.ParameterSpec{
				Type: func() any {
					// Use Type if available, otherwise Types for complex types
					if prop.Type != "" {
						return prop.Type
					} else if len(prop.Types) > 0 {
						return prop.Types
					}
					return "string" // default to string if type is not specified
				}(),
				Description: prop.Description,
			}

			if prop.Type == "array" {
				paramSpec.Items = extractItemsFromSchema(prop.Items)
			}

			if len(prop.Enum) > 0 {
				paramSpec.Enum = append(paramSpec.Enum, prop.Enum...)
			}

			if slices.Contains(toolSchema.Required, name) {
				paramSpec.Required = true
			}
			params[name] = paramSpec
		}
	}
	return params
}

// asString returns v as a string if it is one, empty string otherwise. Used
// to inspect JSON Schema "type" fields, which may be a plain string or a
// union like []string{"array","null"}; union handling is left to
// LazyMCPTool.Parameters which has more context.
func asString(v interface{}) string {
	s, _ := v.(string)
	return s
}

// extractItemsFromMap derives a ParameterSpec for an array's items from the
// raw JSON Schema fragment at `items`. Falls back to {Type:"string"} when
// the fragment is missing or malformed so Gemini/OpenAI accept the
// resulting function declaration.
func extractItemsFromMap(raw interface{}) *contracts.ParameterSpec {
	defaultSpec := &contracts.ParameterSpec{Type: "string"}
	itemsMap, ok := raw.(map[string]interface{})
	if !ok {
		return defaultSpec
	}
	itemType, _ := itemsMap["type"].(string)
	if itemType == "" {
		itemType = "string"
	}
	spec := &contracts.ParameterSpec{Type: itemType}
	if enum, ok := itemsMap["enum"].([]interface{}); ok {
		spec.Enum = enum
	}
	return spec
}

// extractItemsFromSchema is the *jsonschema.Schema analogue of
// extractItemsFromMap.
func extractItemsFromSchema(items *jsonschema.Schema) *contracts.ParameterSpec {
	defaultSpec := &contracts.ParameterSpec{Type: "string"}
	if items == nil {
		return defaultSpec
	}
	itemType := items.Type
	if itemType == "" {
		itemType = "string"
	}
	spec := &contracts.ParameterSpec{Type: itemType}
	if len(items.Enum) > 0 {
		spec.Enum = append(spec.Enum, items.Enum...)
	}
	return spec
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args string) (string, error) {
	// This is the same as Run for Tool
	return t.Run(ctx, args)
}
