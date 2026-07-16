package agent

import (
	"fmt"

	agentconfig "nu/internal/agent/config"
	"nu/internal/contracts"
	"nu/internal/llm/azureopenai"

	// createOpenAIClient creates an OpenAI LLM client
	"nu/internal/llm/deepseek"
	"nu/internal/llm/openai"
)

func createOpenAIClient(config *agentconfig.LLMProviderYAML) (contracts.LLM, error) {
	var options []openai.Option

	// Get API key from config or environment
	apiKey := getConfigString(config.Config, "api_key")
	if apiKey == "" {
		apiKey = agentconfig.GetEnvValue("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for OpenAI provider (set OPENAI_API_KEY or provide in config)")
	}

	// Set model - use config model or fallback to OPENAI_MODEL env var
	model := agentconfig.ExpandEnv(config.Model)
	if model == "" {
		model = getConfigString(config.Config, "model")
	}
	if model == "" {
		model = agentconfig.GetEnvValue("OPENAI_MODEL")
	}
	if model != "" {
		options = append(options, openai.WithModel(model))
	}

	// Set base URL if provided (for custom endpoints)
	if baseURL := getConfigString(config.Config, "base_url"); baseURL != "" {
		options = append(options, openai.WithBaseURL(baseURL))
	}

	return openai.NewClient(apiKey, options...), nil
}

// createDeepSeekClient creates a DeepSeek LLM client
func createDeepSeekClient(config *agentconfig.LLMProviderYAML) (contracts.LLM, error) {
	var options []deepseek.Option

	// Get API key from config or environment
	apiKey := getConfigString(config.Config, "api_key")
	if apiKey == "" {
		// Fallback to DEEPSEEK_API_KEY environment variable
		apiKey = agentconfig.GetEnvValue("DEEPSEEK_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for DeepSeek provider (set DEEPSEEK_API_KEY or provide in config)")
	}

	// Set model - use config model or fallback to DEEPSEEK_MODEL env var
	model := agentconfig.ExpandEnv(config.Model)
	if model == "" {
		model = getConfigString(config.Config, "model")
	}
	if model == "" {
		model = agentconfig.GetEnvValue("DEEPSEEK_MODEL")
	}
	if model != "" {
		options = append(options, deepseek.WithModel(model))
	}

	// Set base URL if provided (for custom endpoints)
	if baseURL := getConfigString(config.Config, "base_url"); baseURL != "" {
		options = append(options, deepseek.WithBaseURL(baseURL))
	}

	return deepseek.NewClient(apiKey, options...), nil
}

// createAzureOpenAIClient creates an Azure OpenAI LLM client
func createAzureOpenAIClient(config *agentconfig.LLMProviderYAML) (contracts.LLM, error) {
	var options []azureopenai.Option

	// Get API key from config or environment
	apiKey := getConfigString(config.Config, "api_key")
	if apiKey == "" {
		apiKey = agentconfig.GetEnvValue("AZURE_OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for Azure OpenAI provider (set AZURE_OPENAI_API_KEY or provide in config)")
	}

	// Get required endpoint
	endpoint := getConfigString(config.Config, "endpoint")
	if endpoint == "" {
		endpoint = agentconfig.GetEnvValue("AZURE_OPENAI_ENDPOINT")
	}
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required for Azure OpenAI provider (set AZURE_OPENAI_ENDPOINT or provide in config)")
	}

	// Get required deployment name
	deployment := getConfigString(config.Config, "deployment")
	if deployment == "" {
		deployment = agentconfig.ExpandEnv(config.Model)
	}
	if deployment == "" {
		deployment = agentconfig.GetEnvValue("AZURE_OPENAI_DEPLOYMENT")
	}
	if deployment == "" {
		return nil, fmt.Errorf("deployment is required for Azure OpenAI provider (set AZURE_OPENAI_DEPLOYMENT or provide in config)")
	}

	// Get API version
	apiVersion := getConfigString(config.Config, "api_version")
	if apiVersion == "" {
		apiVersion = agentconfig.GetEnvValue("AZURE_OPENAI_API_VERSION")
	}
	if apiVersion == "" {
		apiVersion = "2024-02-01" // Default API version
	}

	options = append(options, azureopenai.WithDeployment(deployment))
	options = append(options, azureopenai.WithAPIVersion(apiVersion))

	return azureopenai.NewClient(apiKey, endpoint, deployment, options...), nil
}
