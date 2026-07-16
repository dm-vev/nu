package config

import "time"

// Config represents the global configuration for the Agent SDK
type Config struct {
	// LLM configuration
	LLM struct {
		// OpenAI configuration
		OpenAI struct {
			APIKey         string
			Model          string
			Temperature    float64
			BaseURL        string
			Timeout        time.Duration
			EmbeddingModel string
		}

		// Anthropic configuration
		Anthropic struct {
			APIKey      string
			Model       string
			Temperature float64
			BaseURL     string
			Timeout     time.Duration
		}

		// Azure OpenAI configuration
		AzureOpenAI struct {
			APIKey       string
			Temperature  float64
			BaseURL      string
			Region       string
			ResourceName string
			Deployment   string
			APIVersion   string
			Timeout      time.Duration
		}
	}

	// Memory configuration
	Memory struct {
		// Redis configuration
		Redis struct {
			URL      string
			Password string
			DB       int
		}
	}

	// VectorStore configuration
	VectorStore struct {
		// Weaviate configuration
		Weaviate struct {
			URL       string
			APIKey    string
			Scheme    string
			Host      string
			ClassName string
		}
	}

	// DataStore configuration
	DataStore struct {
		// Supabase configuration
		Supabase struct {
			URL    string
			APIKey string
			Table  string
		}
	}

	// Tools configuration
	Tools struct {
		// Web search configuration
		WebSearch struct {
			GoogleAPIKey         string
			GoogleSearchEngineID string
		}
		// GitHub configuration
		GitHub struct {
			Token string
		}
	}

	// Tracing configuration
	Tracing struct {
		// Langfuse configuration
		Langfuse struct {
			Enabled        bool
			SecretKey      string
			PublicKey      string
			Host           string
			Environment    string
			IncludeContent bool
		}

		// OpenTelemetry configuration
		OpenTelemetry struct {
			Enabled           bool
			ServiceName       string
			CollectorEndpoint string
		}
	}

	// Multitenancy configuration
	Multitenancy struct {
		Enabled      bool
		DefaultOrgID string
	}

	// Guardrails configuration
	Guardrails struct {
		Enabled    bool
		ConfigPath string
	}

	// ConfigService configuration
	ConfigService struct {
		Host string
	}
}

// OpenAIConfig contains OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey      string
	Model       string
	Temperature float64
	BaseURL     string
	Timeout     time.Duration
}

// AnthropicConfig contains Anthropic-specific configuration
type AnthropicConfig struct {
	APIKey      string
	Model       string
	Temperature float64
	BaseURL     string
	Timeout     time.Duration
}

// AzureOpenAIConfig contains Azure OpenAI-specific configuration
type AzureOpenAIConfig struct {
	APIKey       string
	Temperature  float64
	BaseURL      string
	Region       string
	ResourceName string
	Deployment   string
	APIVersion   string
	Timeout      time.Duration
}
