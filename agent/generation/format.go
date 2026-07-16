package generation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// IsStructuredJSONResponse checks if message content is a structured JSON response.
func IsStructuredJSONResponse(content string) bool {
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
}

// ConvertToHumanReadable converts a JSON response to a concise history summary.
func ConvertToHumanReadable(jsonContent string) string {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &jsonMap); err != nil {
		return "[Generated structured response]"
	}

	var parts []string
	for key, value := range jsonMap {
		switch v := value.(type) {
		case string:
			if v != "" && v != "null" {
				parts = append(parts, fmt.Sprintf("%s: %s", key, v))
			}
		case []interface{}:
			if len(v) > 0 {
				if str, ok := v[0].(string); ok && str != "" {
					parts = append(parts, fmt.Sprintf("%s: %s", key, str))
				}
			}
		case bool:
			parts = append(parts, fmt.Sprintf("%s: %t", key, v))
		case float64, int:
			parts = append(parts, fmt.Sprintf("%s: %v", key, v))
		}
	}

	if len(parts) == 0 {
		return "[Generated structured response]"
	}
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return "[AI: " + strings.Join(parts, ", ") + "]"
}
