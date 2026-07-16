package anthropic

import "strings"

// anthropicCreateExampleFromSchema creates an example JSON structure based on the schema
func anthropicCreateExampleFromSchema(schema map[string]interface{}) map[string]interface{} {
	example := make(map[string]interface{})
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for key, value := range properties {
			if prop, ok := value.(map[string]interface{}); ok {
				example[key] = anthropicGetExampleValue(prop)
			}
		}
	}
	return example
}

func anthropicGetExampleValue(prop map[string]interface{}) interface{} {
	propType, _ := prop["type"].(string)
	description, _ := prop["description"].(string)
	switch propType {
	case "string":
		if description != "" {
			return "example_" + strings.ToLower(strings.ReplaceAll(description, " ", "_"))[:20]
		}
		return "example_string"
	case "number":
		return 42.5
	case "integer":
		return 42
	case "boolean":
		return true
	case "array":
		if items, ok := prop["items"].(map[string]interface{}); ok {
			if itemType, ok := items["type"].(string); ok {
				switch itemType {
				case "string":
					return []string{"example_item_1", "example_item_2"}
				case "number", "integer":
					return []int{1, 2, 3}
				case "object":
					exampleObj := anthropicGetExampleValue(items)
					return []interface{}{exampleObj, exampleObj}
				}
			}
		}
		return []interface{}{"item1", "item2"}
	case "object":
		if properties, ok := prop["properties"].(map[string]interface{}); ok {
			obj := make(map[string]interface{})
			for k, v := range properties {
				if subProp, ok := v.(map[string]interface{}); ok {
					obj[k] = anthropicGetExampleValue(subProp)
				}
			}
			return obj
		}
		return map[string]interface{}{"key": "value"}
	default:
		return "example_value"
	}
}

// anthropicExtractJSONFromResponse extracts JSON content from a response that may contain markdown or explanatory text
func anthropicExtractJSONFromResponse(response string) string {
	jsonStart := strings.Index(response, "```json")
	if jsonStart >= 0 {
		jsonStart += len("```json")
		jsonEnd := strings.Index(response[jsonStart:], "```")
		if jsonEnd > 0 {
			return strings.TrimSpace(response[jsonStart : jsonStart+jsonEnd])
		}
	}

	jsonStart = strings.Index(response, "```")
	if jsonStart >= 0 {
		jsonStart += len("```")
		contentAfterMarker := response[jsonStart:]
		newlineIdx := strings.Index(contentAfterMarker, "\n")
		if newlineIdx >= 0 {
			contentAfterMarker = contentAfterMarker[newlineIdx+1:]
		}
		jsonEnd := strings.Index(contentAfterMarker, "```")
		if jsonEnd > 0 {
			extracted := strings.TrimSpace(contentAfterMarker[:jsonEnd])
			if anthropicIsValidJSONStart(extracted) {
				return extracted
			}
		}
	}

	jsonStart = strings.Index(response, "{")
	if jsonStart >= 0 {
		braceCount := 0
		inString := false
		escapeNext := false
		for i := jsonStart; i < len(response); i++ {
			char := response[i]
			if escapeNext {
				escapeNext = false
				continue
			}
			if char == '\\' {
				escapeNext = true
				continue
			}
			if char == '"' {
				inString = !inString
				continue
			}
			if !inString {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
					if braceCount == 0 {
						extracted := strings.TrimSpace(response[jsonStart : i+1])
						if anthropicIsValidJSONStart(extracted) {
							return extracted
						}
						break
					}
				}
			}
		}
	}
	return response
}

func anthropicIsValidJSONStart(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}
