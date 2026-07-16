package anthropic

import "strings"

// ModelName constants for supported Anthropic models
const (
	Claude35Haiku  = "claude-3-5-haiku-latest"
	Claude35Sonnet = "claude-3-5-sonnet-latest"
	Claude3Opus    = "claude-3-opus-latest"
	Claude37Sonnet = "claude-3-7-sonnet-20250219"
	ClaudeSonnet4  = "claude-sonnet-4-20250514"
	ClaudeSonnet45 = "claude-sonnet-4-5-20250929"
	ClaudeOpus4    = "claude-opus-4-20250514"
	ClaudeOpus41   = "claude-opus-4-1-20250805"
	ClaudeOpus45   = "claude-opus-4-5-20251101"

	BedrockClaude35Haiku  = "anthropic.claude-3-5-haiku-20241022-v1:0"
	BedrockClaude35Sonnet = "anthropic.claude-3-5-sonnet-20241022-v2:0"
	BedrockClaude3Opus    = "anthropic.claude-3-opus-20240229-v1:0"
	BedrockClaude37Sonnet = "anthropic.claude-3-7-sonnet-20250219-v1:0"
	BedrockClaudeSonnet4  = "anthropic.claude-sonnet-4-20250514-v1:0"
	BedrockClaudeSonnet45 = "anthropic.claude-sonnet-4-5-20250929-v1:0"
	BedrockClaudeOpus4    = "anthropic.claude-opus-4-20250514-v1:0"
	BedrockClaudeOpus41   = "anthropic.claude-opus-4-1-20250805-v1:0"
	BedrockClaudeOpus45   = "anthropic.claude-opus-4-5-20251101-v1:0"
)

// AnthropicSupportsThinking returns true if the model supports thinking tokens
func SupportsThinking(model string) bool {
	normalizedModel := model
	if strings.Contains(model, ".anthropic.claude") {
		parts := strings.SplitN(model, ".anthropic.", 2)
		if len(parts) == 2 {
			normalizedModel = parts[1]
		}
	} else if strings.HasPrefix(model, "anthropic.claude") {
		normalizedModel = strings.TrimPrefix(model, "anthropic.")
	}
	supportedModels := []string{
		"claude-3-7-sonnet-20250219", "claude-sonnet-4-20250514", "claude-sonnet-4-5-20250929",
		"claude-opus-4-20250514", "claude-opus-4-1-20250805", "claude-opus-4-5-20251101",
		"claude-sonnet-4@20250514", "claude-sonnet-4-v1@20250514", "claude-sonnet-4-5@20250929",
		"claude-opus-4@20250514", "claude-opus-4-v1@20250514", "claude-opus-4-1@20250805", "claude-opus-4-5@20251101",
		"claude-3-7-sonnet-20250219-v1:0", "claude-sonnet-4-20250514-v1:0", "claude-sonnet-4-5-20250929-v1:0",
		"claude-opus-4-20250514-v1:0", "claude-opus-4-1-20250805-v1:0", "claude-opus-4-5-20251101-v1:0",
	}
	for _, supportedModel := range supportedModels {
		if normalizedModel == supportedModel || model == supportedModel {
			return true
		}
	}
	return false
}
