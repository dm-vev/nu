package anthropic

import "strings"

// IsVertexModel checks if a model name is in Vertex AI format (contains @)
func IsVertexModel(model string) bool {
	return strings.Contains(model, "@")
}

// ConvertToVertexModel converts a standard Anthropic model name to Vertex AI format
// This is a basic mapping - users should use the correct Vertex model names
func ConvertToVertexModel(model string) string {
	switch model {
	case Claude35Haiku:
		return "claude-3-5-haiku@20241022"
	case Claude35Sonnet:
		return "claude-3-5-sonnet-v2@20241022"
	case Claude3Opus:
		return "claude-3-opus@20240229"
	case Claude37Sonnet:
		return "claude-3-7-sonnet@20250219"
	case ClaudeSonnet4:
		return "claude-sonnet-4-v1@20250514"
	case ClaudeOpus4:
		return "claude-opus-4-v1@20250514"
	case ClaudeOpus41:
		return "claude-opus-4-1@20250805"
	default:
		return model
	}
}
