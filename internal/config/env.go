package config

import (
	"os"
	"strconv"
	"time"
)

// initLLMConfig initializes LLM configuration with defaults
func initLLMConfig(config *Config) {
	// OpenAI defaults
	config.LLM.OpenAI.APIKey = getEnvString("OPENAI_API_KEY", "")
	config.LLM.OpenAI.Model = getEnvString("OPENAI_MODEL", "gpt-4o-mini")
	config.LLM.OpenAI.Temperature = getEnvFloat("OPENAI_TEMPERATURE", 0.7)
	config.LLM.OpenAI.BaseURL = getEnvString("OPENAI_BASE_URL", "")
	config.LLM.OpenAI.Timeout = time.Duration(getEnvInt("OPENAI_TIMEOUT", 60)) * time.Second

	// Anthropic defaults
	config.LLM.Anthropic.APIKey = getEnvString("ANTHROPIC_API_KEY", "")
	config.LLM.Anthropic.Model = getEnvString("ANTHROPIC_MODEL", "claude-3-7-sonnet-20240307")
	config.LLM.Anthropic.Temperature = getEnvFloat("ANTHROPIC_TEMPERATURE", 0.7)
	config.LLM.Anthropic.BaseURL = getEnvString("ANTHROPIC_BASE_URL", "")
	config.LLM.Anthropic.Timeout = time.Duration(getEnvInt("ANTHROPIC_TIMEOUT", 60)) * time.Second

	// Azure OpenAI defaults
	config.LLM.AzureOpenAI.APIKey = getEnvString("AZURE_OPENAI_API_KEY", "")
	config.LLM.AzureOpenAI.Temperature = getEnvFloat("AZURE_OPENAI_TEMPERATURE", 0.7)
	config.LLM.AzureOpenAI.BaseURL = getEnvString("AZURE_OPENAI_BASE_URL", "")
	config.LLM.AzureOpenAI.Region = getEnvString("AZURE_OPENAI_REGION", "")
	config.LLM.AzureOpenAI.ResourceName = getEnvString("AZURE_OPENAI_RESOURCE_NAME", "")
	config.LLM.AzureOpenAI.Deployment = getEnvString("AZURE_OPENAI_DEPLOYMENT", "")
	config.LLM.AzureOpenAI.APIVersion = getEnvString("AZURE_OPENAI_API_VERSION", "2024-08-01-preview")
	config.LLM.AzureOpenAI.Timeout = time.Duration(getEnvInt("AZURE_OPENAI_TIMEOUT", 60)) * time.Second
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// getEnvFloat gets a float environment variable or returns a default value
func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// getEnvString gets a string environment variable or returns a default value
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
