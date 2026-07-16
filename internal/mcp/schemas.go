package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// SchemaValidator handles validation of tool outputs against their schemas
type SchemaValidator struct {
	logger telemetry.Logger
}

func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{logger: telemetry.NewLogger()}
}

func (sv *SchemaValidator) ValidateToolResponse(ctx context.Context, tool contracts.MCPTool, response *contracts.MCPToolResponse) error {
	if tool.OutputSchema == nil {
		return nil
	}
	if response.StructuredContent == nil {
		sv.logger.Warn(ctx, "Tool has output schema but response has no structured content", map[string]interface{}{
			"tool_name": tool.Name,
		})
		return nil
	}
	return sv.validateAgainstSchema(ctx, response.StructuredContent, tool.OutputSchema, tool.Name)
}

func (sv *SchemaValidator) validateAgainstSchema(ctx context.Context, data interface{}, schema interface{}, toolName string) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid schema format for tool %s: expected object", toolName)
	}
	schemaType, hasType := schemaMap["type"]
	if !hasType {
		if len(schemaMap) == 0 {
			return nil
		}
		return fmt.Errorf("schema missing 'type' field for tool %s", toolName)
	}
	switch schemaType {
	case "object":
		return sv.validateObject(ctx, data, schemaMap, toolName)
	case "array":
		return sv.validateArray(ctx, data, schemaMap, toolName)
	case "string":
		return sv.validateString(ctx, data, schemaMap, toolName)
	case "number", "integer":
		return sv.validateNumber(ctx, data, schemaMap, toolName)
	case "boolean":
		return sv.validateBoolean(ctx, data, schemaMap, toolName)
	default:
		sv.logger.Warn(ctx, "Unknown schema type", map[string]interface{}{
			"tool_name": toolName, "type": schemaType,
		})
		return nil
	}
}

func (sv *SchemaValidator) validateObject(ctx context.Context, data interface{}, schema map[string]interface{}, toolName string) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object for tool %s, got %T", toolName, data)
	}
	if required, hasRequired := schema["required"]; hasRequired {
		if requiredFields, ok := required.([]interface{}); ok {
			for _, field := range requiredFields {
				if fieldName, ok := field.(string); ok {
					if _, exists := dataMap[fieldName]; !exists {
						return fmt.Errorf("missing required field '%s' in tool %s response", fieldName, toolName)
					}
				}
			}
		}
	}
	if properties, hasProperties := schema["properties"]; hasProperties {
		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			for fieldName, fieldValue := range dataMap {
				if fieldSchema, exists := propertiesMap[fieldName]; exists {
					if err := sv.validateAgainstSchema(ctx, fieldValue, fieldSchema, fmt.Sprintf("%s.%s", toolName, fieldName)); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (sv *SchemaValidator) validateArray(ctx context.Context, data interface{}, schema map[string]interface{}, toolName string) error {
	dataSlice := reflect.ValueOf(data)
	if dataSlice.Kind() != reflect.Slice && dataSlice.Kind() != reflect.Array {
		return fmt.Errorf("expected array for tool %s, got %T", toolName, data)
	}
	if items, hasItems := schema["items"]; hasItems {
		for i := 0; i < dataSlice.Len(); i++ {
			if err := sv.validateAgainstSchema(ctx, dataSlice.Index(i).Interface(), items, fmt.Sprintf("%s[%d]", toolName, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (sv *SchemaValidator) validateString(ctx context.Context, data interface{}, schema map[string]interface{}, toolName string) error {
	if _, ok := data.(string); !ok {
		return fmt.Errorf("expected string for tool %s, got %T", toolName, data)
	}
	return nil
}

func (sv *SchemaValidator) validateNumber(ctx context.Context, data interface{}, schema map[string]interface{}, toolName string) error {
	switch data.(type) {
	case int, int64, float64, float32, json.Number:
		return nil
	default:
		return fmt.Errorf("expected number for tool %s, got %T", toolName, data)
	}
}

func (sv *SchemaValidator) validateBoolean(ctx context.Context, data interface{}, schema map[string]interface{}, toolName string) error {
	if _, ok := data.(bool); !ok {
		return fmt.Errorf("expected boolean for tool %s, got %T", toolName, data)
	}
	return nil
}
