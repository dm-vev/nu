package agent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentconfig "nu/internal/agent/config"
	"nu/internal/contracts"
)

// createLLMFromConfig creates an LLM client from YAML configuration
func createLLMFromConfig(config *agentconfig.LLMProviderYAML) (contracts.LLM, error) {
	if config == nil || config.Provider == "" {
		return nil, fmt.Errorf("LLM provider configuration is required")
	}

	provider := strings.ToLower(config.Provider)

	switch provider {
	case "anthropic":
		return createAnthropicClient(config)
	case "openai":
		return createOpenAIClient(config)
	case "azureopenai", "azure_openai":
		return createAzureOpenAIClient(config)
	case "deepseek":
		return createDeepSeekClient(config)
	case "gemini":
		return createGeminiClient(config)
	case "ollama":
		return createOllamaClient(config)
	case "vllm":
		return createVllmClient(config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: anthropic, openai, azureopenai, deepseek, gemini, ollama, vllm)", provider)
	}
}

// parseGoogleCredentials parses Google Application Credentials from multiple formats.
// It supports three input formats with automatic detection:
//  1. File path: Reads and validates the JSON content from the specified file
//  2. Base64-encoded JSON: Decodes the base64 string and validates the JSON
//  3. Raw JSON string: Uses the content directly after validation
//
// The function validates that the final output is valid JSON before returning.
// Returns the JSON credential content as a string, or an error if parsing fails.
func parseGoogleCredentials(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty credentials input")
	}

	// 1. Check if it's a file path
	cleanPath := filepath.Clean(input)
	if _, err := os.Stat(cleanPath); err == nil { // #nosec G703 - input comes from agent configuration, not untrusted user input
		data, err := os.ReadFile(cleanPath) // #nosec G304 G703 -- File path comes from agent configuration, not untrusted user input
		if err != nil {
			return "", fmt.Errorf("failed to read credentials file: %w", err)
		}
		// Validate it's valid JSON
		if !json.Valid(data) {
			return "", fmt.Errorf("credentials file does not contain valid JSON")
		}
		return string(data), nil
	}

	// 2. Check if it's base64 encoded
	if decoded, err := base64.StdEncoding.DecodeString(input); err == nil {
		// Validate it's valid JSON
		if json.Valid(decoded) {
			return string(decoded), nil
		}
	}

	// 3. Treat as raw JSON content
	if json.Valid([]byte(input)) {
		return input, nil
	}

	return "", fmt.Errorf("invalid credentials format: not a valid file path, base64-encoded JSON, or raw JSON content")
}

// Helper function to extract string values from config map
func getConfigString(config map[string]interface{}, key string) string {
	if config == nil {
		return ""
	}
	if value, exists := config[key]; exists {
		if str, ok := value.(string); ok {
			// Expand environment variables using SDK's ExpandEnv that supports .env files
			return agentconfig.ExpandEnv(str)
		}
	}
	return ""
}
