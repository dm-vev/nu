package llm

import "strings"

// extractJSON extracts JSON from a string, handling markdown code blocks
func extractJSON(s string) string {
	// First, try to extract from markdown code blocks
	if strings.Contains(s, "```json") {
		start := strings.Index(s, "```json")
		if start != -1 {
			start += 7 // Move past "```json"
			end := strings.Index(s[start:], "```")
			if end != -1 {
				jsonContent := strings.TrimSpace(s[start : start+end])
				return jsonContent
			}
		}
	}

	// Also handle generic code blocks that might contain JSON
	if strings.Contains(s, "```") {
		parts := strings.Split(s, "```")
		for i := 1; i < len(parts); i += 2 { // Check odd indices (inside code blocks)
			content := strings.TrimSpace(parts[i])
			// Remove language identifier if present (e.g., "json\n{...}")
			if strings.Contains(content, "\n") && strings.HasPrefix(content, "json") {
				content = strings.TrimSpace(content[4:])
			}
			// Check if this looks like JSON
			if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
				return content
			}
		}
	}

	// Fallback to original logic - find the first { and the last }
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start == -1 || end == -1 || end <= start {
		return ""
	}

	return s[start : end+1]
}
