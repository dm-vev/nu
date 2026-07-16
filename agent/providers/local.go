package providers

import (
	"fmt"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm/vllm"

	// createOllamaClient creates an Ollama LLM client
	"github.com/dm-vev/nu/internal/llm/ollama"
)

func createOllamaClient(cfg *config.LLMProviderYAML) (contracts.LLM, error) {
	var options []ollama.Option

	// Get base URL from config or environment, default to localhost
	baseURL := getConfigString(cfg.Config, "base_url")
	if baseURL == "" {
		baseURL = config.GetEnvValue("OLLAMA_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama URL
	}

	// Set base URL
	options = append(options, ollama.WithBaseURL(baseURL))

	// Set model - use config model or fallback to OLLAMA_MODEL env var
	model := config.ExpandEnv(cfg.Model)
	if model == "" {
		model = getConfigString(cfg.Config, "model")
	}
	if model == "" {
		model = config.GetEnvValue("OLLAMA_MODEL")
	}
	if model != "" {
		options = append(options, ollama.WithModel(model))
	}

	return ollama.NewClient(options...), nil
}

// createVllmClient creates a vLLM LLM client
func createVllmClient(cfg *config.LLMProviderYAML) (contracts.LLM, error) {
	var options []vllm.Option

	// Get base URL from config or environment
	baseURL := getConfigString(cfg.Config, "base_url")
	if baseURL == "" {
		baseURL = config.GetEnvValue("VLLM_BASE_URL")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required for vLLM provider (set VLLM_BASE_URL or provide in config)")
	}

	// Set base URL
	options = append(options, vllm.WithBaseURL(baseURL))

	// Set model - use config model or fallback to VLLM_MODEL env var
	model := config.ExpandEnv(cfg.Model)
	if model == "" {
		model = getConfigString(cfg.Config, "model")
	}
	if model == "" {
		model = config.GetEnvValue("VLLM_MODEL")
	}
	if model != "" {
		options = append(options, vllm.WithModel(model))
	}

	return vllm.NewClient(options...), nil
}
