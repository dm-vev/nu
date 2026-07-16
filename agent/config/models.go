package config

import "time"

// ConfigSourceMetadata tracks where a configuration was loaded from
type ConfigSourceMetadata struct {
	Type        string            `yaml:"type" json:"type"`     // "local", "remote"
	Source      string            `yaml:"source" json:"source"` // file path or service URL
	AgentID     string            `yaml:"agent_id,omitempty" json:"agent_id,omitempty"`
	AgentName   string            `yaml:"agent_name,omitempty" json:"agent_name,omitempty"`
	Environment string            `yaml:"environment,omitempty" json:"environment,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
	LoadedAt    time.Time         `yaml:"loaded_at" json:"loaded_at"`
}

// ResponseFormatConfig represents the configuration for the response format of an agent or task
type ResponseFormatConfig struct {
	Type             string                 `yaml:"type"`
	SchemaName       string                 `yaml:"schema_name"`
	SchemaDefinition map[string]interface{} `yaml:"schema_definition"`
}

// AgentConfig represents the configuration for an agent loaded from YAML
type AgentConfig struct {
	Role           string                `yaml:"role"`
	Goal           string                `yaml:"goal"`
	Backstory      string                `yaml:"backstory"`
	ResponseFormat *ResponseFormatConfig `yaml:"response_format,omitempty"`
	MCP            *MCPConfiguration     `yaml:"mcp,omitempty"`

	// NEW: Behavioral settings
	MaxIterations       *int  `yaml:"max_iterations,omitempty"`
	DisableFinalSummary *bool `yaml:"disable_final_summary,omitempty"`
	RequirePlanApproval *bool `yaml:"require_plan_approval,omitempty"`

	// NEW: Complex configuration objects
	StreamConfig *StreamConfigYAML `yaml:"stream_config,omitempty"`
	LLMConfig    *LLMConfigYAML    `yaml:"llm_config,omitempty"`
	CacheConfig  *CacheConfigYAML  `yaml:"cache_config,omitempty"`

	// NEW: LLM Provider configuration
	LLMProvider *LLMProviderYAML `yaml:"llm_provider,omitempty"`

	// NEW: Tool configurations
	Tools []ToolConfigYAML `yaml:"tools,omitempty"`

	// NEW: Memory configuration (config only)
	Memory *MemoryConfigYAML `yaml:"memory,omitempty"`

	// NEW: Runtime settings
	Runtime *RuntimeConfigYAML `yaml:"runtime,omitempty"`

	// NEW: Image generation configuration
	ImageGeneration *ImageGenerationYAML `yaml:"image_generation,omitempty"`

	// NEW: Sub-agents configuration (recursive)
	SubAgents map[string]AgentConfig `yaml:"sub_agents,omitempty"`

	// NEW: Configuration source metadata
	ConfigSource *ConfigSourceMetadata `yaml:"config_source,omitempty" json:"config_source,omitempty"`
}

// TaskConfig represents a task definition loaded from YAML
type TaskConfig struct {
	Description    string                `yaml:"description"`
	ExpectedOutput string                `yaml:"expected_output"`
	Agent          string                `yaml:"agent"`
	OutputFile     string                `yaml:"output_file,omitempty"`
	ResponseFormat *ResponseFormatConfig `yaml:"response_format,omitempty"`
}

// StreamConfigYAML represents streaming configuration in YAML
type StreamConfigYAML struct {
	BufferSize                  *int  `yaml:"buffer_size,omitempty"`
	IncludeToolProgress         *bool `yaml:"include_tool_progress,omitempty"`
	IncludeIntermediateMessages *bool `yaml:"include_intermediate_messages,omitempty"`
}

// CacheConfigYAML represents prompt caching configuration in YAML (Anthropic only)
type CacheConfigYAML struct {
	CacheSystemMessage *bool   `yaml:"cache_system_message,omitempty"`
	CacheTools         *bool   `yaml:"cache_tools,omitempty"`
	CacheConversation  *bool   `yaml:"cache_conversation,omitempty"`
	CacheTTL           *string `yaml:"cache_ttl,omitempty"`
}

// LLMConfigYAML represents LLM configuration in YAML
type LLMConfigYAML struct {
	Temperature      *float64 `yaml:"temperature,omitempty"`
	TopP             *float64 `yaml:"top_p,omitempty"`
	FrequencyPenalty *float64 `yaml:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `yaml:"presence_penalty,omitempty"`
	StopSequences    []string `yaml:"stop_sequences,omitempty"`
	EnableReasoning  *bool    `yaml:"enable_reasoning,omitempty"`
	ReasoningBudget  *int     `yaml:"reasoning_budget,omitempty"`
	Reasoning        *string  `yaml:"reasoning,omitempty"`
}

// LLMProviderYAML represents LLM provider configuration in YAML
type LLMProviderYAML struct {
	Provider string                 `yaml:"provider"`
	Model    string                 `yaml:"model,omitempty"`
	Config   map[string]interface{} `yaml:"config,omitempty"`
}

// ToolConfigYAML represents tool configuration in YAML
type ToolConfigYAML struct {
	Type        string                 `yaml:"type"` // "builtin", "custom", "mcp", "agent"
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Config      map[string]interface{} `yaml:"config,omitempty"`
	Enabled     *bool                  `yaml:"enabled,omitempty"`

	// For agent tools
	URL     string `yaml:"url,omitempty"`     // Remote agent URL
	Timeout string `yaml:"timeout,omitempty"` // Timeout duration
}

// MemoryConfigYAML represents memory configuration in YAML
type MemoryConfigYAML struct {
	Type   string                 `yaml:"type"` // "buffer", "redis", "vector"
	Config map[string]interface{} `yaml:"config,omitempty"`
}

// RuntimeConfigYAML represents runtime behavior settings in YAML
type RuntimeConfigYAML struct {
	LogLevel        string `yaml:"log_level,omitempty"` // "debug", "info", "warn", "error"
	EnableTracing   *bool  `yaml:"enable_tracing,omitempty"`
	EnableMetrics   *bool  `yaml:"enable_metrics,omitempty"`
	TimeoutDuration string `yaml:"timeout_duration,omitempty"` // "30s", "5m"
}

// ImageGenerationYAML represents image generation configuration in YAML
type ImageGenerationYAML struct {
	Enabled          *bool                  `yaml:"enabled,omitempty"`
	Provider         string                 `yaml:"provider,omitempty"` // "gemini"
	Model            string                 `yaml:"model,omitempty"`    // e.g., "gemini-2.5-flash-image"
	Config           map[string]interface{} `yaml:"config,omitempty"`
	Storage          *ImageStorageYAML      `yaml:"storage,omitempty"`
	MultiTurnEditing *MultiTurnEditingYAML  `yaml:"multi_turn_editing,omitempty"`
}

// MultiTurnEditingYAML represents multi-turn image editing configuration in YAML
type MultiTurnEditingYAML struct {
	Enabled           *bool  `yaml:"enabled,omitempty"`
	Model             string `yaml:"model,omitempty"`           // e.g., "gemini-3-pro-image-preview"
	SessionTimeout    string `yaml:"session_timeout,omitempty"` // e.g., "30m"
	MaxSessionsPerOrg *int   `yaml:"max_sessions_per_org,omitempty"`
}

// ImageStorageYAML represents image storage configuration in YAML
type ImageStorageYAML struct {
	Type  string            `yaml:"type,omitempty"` // "local", "gcs"
	Local *LocalStorageYAML `yaml:"local,omitempty"`
	GCS   *GCSStorageYAML   `yaml:"gcs,omitempty"`
}

// LocalStorageYAML represents local storage configuration in YAML
type LocalStorageYAML struct {
	Path    string `yaml:"path,omitempty"`
	BaseURL string `yaml:"base_url,omitempty"`
}

// GCSStorageYAML represents GCS storage configuration in YAML
type GCSStorageYAML struct {
	Bucket              string `yaml:"bucket,omitempty"`
	Prefix              string `yaml:"prefix,omitempty"`
	CredentialsFile     string `yaml:"credentials_file,omitempty"`
	CredentialsJSON     string `yaml:"credentials_json,omitempty"`
	SignedURLExpiration string `yaml:"signed_url_expiration,omitempty"`
}

// AgentConfigs represents a map of agent configurations
type AgentConfigs map[string]AgentConfig

// TaskConfigs represents a map of task configurations
type TaskConfigs map[string]TaskConfig
