package agent

import (
	"context"
	"time"

	agentconfig "nu/internal/agent/config"
	"nu/internal/agent/plans"
	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// LazyMCPConfig holds configuration for lazy MCP server initialization
type LazyMCPConfig struct {
	Name              string
	Type              string // "stdio" or "http"
	Command           string
	Args              []string
	Env               []string
	URL               string
	Token             string // Bearer token for HTTP authentication
	Tools             []LazyMCPToolConfig
	HttpTransportMode string   // "sse" or "streamable"
	AllowedTools      []string // List of allowed tool names for this MCP server
}

// LazyMCPToolConfig holds configuration for a lazy MCP tool
type LazyMCPToolConfig struct {
	Name        string
	Description string
	Schema      interface{}
}

// CustomRunFunction represents a custom function that can replace the default Run behavior
type CustomRunFunction func(ctx context.Context, input string, agent *Agent) (string, error)

// CustomRunStreamFunction represents a custom function that can replace the default RunStream behavior
type CustomRunStreamFunction func(ctx context.Context, input string, agent *Agent) (<-chan contracts.AgentStreamEvent, error)

// Agent represents an AI agent
type Agent struct {
	llm                  contracts.LLM
	memory               contracts.Memory
	datastore            contracts.DataStore     // DataStore for persistent data storage (PostgreSQL, Supabase, etc.)
	graphRAGStore        contracts.GraphRAGStore // GraphRAG store for knowledge graph operations
	tools                []contracts.Tool
	subAgents            []*Agent // Sub-agents that can be called as tools
	orgID                string
	tracer               contracts.Tracer
	guardrails           contracts.Guardrails
	logger               telemetry.Logger // Logger for the agent
	systemPrompt         string
	name                 string                               // Name of the agent, e.g., "PlatformOps", "Math", "Research"
	description          string                               // Description of what the agent does
	requirePlanApproval  bool                                 // New field to control whether execution plans require approval
	planStore            *plans.ExecutionPlanStore            // Store for execution plans
	planGenerator        *plans.ExecutionPlanGeneratorService // Generator for execution plans
	planExecutor         *plans.ExecutionPlanExecutor         // Executor for execution plans
	generatedAgentConfig *agentconfig.AgentConfig
	generatedTaskConfigs agentconfig.TaskConfigs
	responseFormat       *contracts.ResponseFormat // Response format for the agent
	llmConfig            *contracts.LLMConfig
	mcpServers           []contracts.MCPServer        // MCP servers for the agent
	lazyMCPConfigs       []LazyMCPConfig              // Lazy MCP server configurations
	configuredTools      []agentconfig.ToolConfigYAML // Tools applied after all options are known
	maxIterations        int                          // Maximum number of tool-calling iterations (default: 2)
	disableFinalSummary  bool                         // When true, skip the final summary LLM call
	streamConfig         *contracts.StreamConfig      // Streaming configuration for the agent
	cacheConfig          *contracts.CacheConfig       // Prompt caching configuration (Anthropic only)

	// Runtime configuration fields
	memoryConfig   map[string]interface{} // Memory configuration from YAML
	timeout        time.Duration          // Agent timeout from runtime config
	tracingEnabled bool                   // Whether tracing is enabled
	metricsEnabled bool                   // Whether metrics are enabled

	// Remote agent fields
	isRemote            bool                        // Whether this is a remote agent
	remoteURL           string                      // URL of the remote agent service
	remoteTimeout       time.Duration               // Timeout for remote agent operations
	remoteClient        contracts.RemoteAgentClient // Injected client for remote communication
	remoteClientFactory func(string) contracts.RemoteAgentClient

	// Custom function fields
	customRunFunc       CustomRunFunction       // Custom run function to replace default behavior
	customRunStreamFunc CustomRunStreamFunction // Custom stream function to replace default streaming behavior
}

// Option represents an option for configuring an agent
type Option func(*Agent)
