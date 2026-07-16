package azureopenai

import "github.com/dm-vev/nu/contracts"

// convertToOpenAISchema converts tool parameters to OpenAI function schema
func (c *Client) convertToOpenAISchema(params map[string]contracts.ParameterSpec) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}
	for name, param := range params {
		property := map[string]interface{}{"type": param.Type, "description": param.Description}
		if param.Default != nil {
			property["default"] = param.Default
		}
		if param.Items != nil {
			property["items"] = map[string]interface{}{"type": param.Items.Type}
			if param.Items.Enum != nil {
				property["items"].(map[string]interface{})["enum"] = param.Items.Enum
			}
		}
		if param.Enum != nil {
			property["enum"] = param.Enum
		}
		properties[name] = property
		if param.Required {
			required = append(required, name)
		}
	}
	return map[string]interface{}{"type": "object", "properties": properties, "required": required}
}
