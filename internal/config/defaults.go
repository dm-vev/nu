package config

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := &Config{}

	// LLM configuration
	initLLMConfig(config)

	// Memory configuration
	config.Memory.Redis.URL = getEnv("REDIS_URL", "localhost:6379")
	config.Memory.Redis.Password = getEnv("REDIS_PASSWORD", "")
	config.Memory.Redis.DB = getEnvInt("REDIS_DB", 0)

	// VectorStore configuration
	config.VectorStore.Weaviate.URL = getEnv("WEAVIATE_URL", "")
	config.VectorStore.Weaviate.APIKey = getEnv("WEAVIATE_API_KEY", "")
	config.VectorStore.Weaviate.Scheme = getEnv("WEAVIATE_SCHEME", "https")
	config.VectorStore.Weaviate.Host = getEnv("WEAVIATE_HOST", "localhost:8080")
	config.VectorStore.Weaviate.ClassName = getEnv("WEAVIATE_CLASS_NAME", "Document")

	// DataStore configuration
	config.DataStore.Supabase.URL = getEnv("SUPABASE_URL", "")
	config.DataStore.Supabase.APIKey = getEnv("SUPABASE_API_KEY", "")
	config.DataStore.Supabase.Table = getEnv("SUPABASE_TABLE", "documents")

	// Tools configuration
	config.Tools.WebSearch.GoogleAPIKey = getEnv("GOOGLE_API_KEY", "")
	config.Tools.WebSearch.GoogleSearchEngineID = getEnv("GOOGLE_SEARCH_ENGINE_ID", "")

	config.Tools.GitHub.Token = getEnv("GITHUB_TOKEN", "")

	// Tracing configuration
	config.Tracing.Langfuse.Enabled = getEnvBool("LANGFUSE_ENABLED", false)
	config.Tracing.Langfuse.SecretKey = getEnv("LANGFUSE_SECRET_KEY", "")
	config.Tracing.Langfuse.PublicKey = getEnv("LANGFUSE_PUBLIC_KEY", "")
	config.Tracing.Langfuse.Host = getEnv("LANGFUSE_HOST", "https://cloud.langfuse.com")
	config.Tracing.Langfuse.Environment = getEnv("LANGFUSE_ENVIRONMENT", "development")
	config.Tracing.Langfuse.IncludeContent = getEnvBool("LANGFUSE_INCLUDE_CONTENT", false)

	config.Tracing.OpenTelemetry.Enabled = getEnvBool("OTEL_ENABLED", false)
	config.Tracing.OpenTelemetry.ServiceName = getEnv("OTEL_SERVICE_NAME", "agent-sdk")
	config.Tracing.OpenTelemetry.CollectorEndpoint = getEnv("OTEL_COLLECTOR_ENDPOINT", "localhost:4317")

	// Multitenancy configuration
	config.Multitenancy.Enabled = getEnvBool("MULTITENANCY_ENABLED", false)
	config.Multitenancy.DefaultOrgID = getEnv("DEFAULT_ORG_ID", "default")

	// Guardrails configuration
	config.Guardrails.Enabled = getEnvBool("GUARDRAILS_ENABLED", false)
	config.Guardrails.ConfigPath = getEnv("GUARDRAILS_CONFIG_PATH", "")

	// ConfigService configuration
	config.ConfigService.Host = getEnv("STAROPS_CONFIG_SERVICE_HOST", "http://starops-config-service-service.starops-config-service.svc.cluster.local:8080")

	return config
}
