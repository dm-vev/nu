package agent

import (
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// WithLLM sets the LLM for the agent.
func WithLLM(llm contracts.LLM) Option { return func(a *Agent) { a.llm = llm } }

// WithMemory sets the memory for the agent.
func WithMemory(memory contracts.Memory) Option { return func(a *Agent) { a.memory = memory } }

// WithDataStore sets the datastore for the agent.
func WithDataStore(datastore contracts.DataStore) Option {
	return func(a *Agent) { a.datastore = datastore }
}

// WithTools appends tools to the agent's tool list.
func WithTools(tools ...contracts.Tool) Option {
	return func(a *Agent) { a.tools = deduplicateTools(append(a.tools, tools...)) }
}

func deduplicateTools(tools []contracts.Tool) []contracts.Tool {
	if len(tools) == 0 {
		return tools
	}
	seen := make(map[string]bool, len(tools))
	result := make([]contracts.Tool, 0, len(tools))
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		name := tool.Name()
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, tool)
	}
	return result
}

// WithOrgID sets the organization ID for multi-tenancy.
func WithOrgID(orgID string) Option { return func(a *Agent) { a.orgID = orgID } }

// WithTracer sets the tracer for the agent.
func WithTracer(tracer contracts.Tracer) Option { return func(a *Agent) { a.tracer = tracer } }

// WithLogger sets the logger for the agent.
func WithLogger(logger telemetry.Logger) Option { return func(a *Agent) { a.logger = logger } }

// WithGuardrails sets the guardrails for the agent.
func WithGuardrails(guardrails contracts.Guardrails) Option {
	return func(a *Agent) { a.guardrails = guardrails }
}

// WithSystemPrompt sets the system prompt for the agent.
func WithSystemPrompt(prompt string) Option { return func(a *Agent) { a.systemPrompt = prompt } }

// WithRequirePlanApproval sets whether execution plans require user approval.
func WithRequirePlanApproval(require bool) Option {
	return func(a *Agent) { a.requirePlanApproval = require }
}

// WithName sets the name for the agent.
func WithName(name string) Option { return func(a *Agent) { a.name = name } }

// WithDescription sets the description for the agent.
func WithDescription(description string) Option {
	return func(a *Agent) { a.description = description }
}

// WithResponseFormat sets the response format for the agent.
func WithResponseFormat(formatType contracts.ResponseFormat) Option {
	return func(a *Agent) { a.responseFormat = &formatType }
}

// WithLLMConfig sets provider-specific LLM configuration.
func WithLLMConfig(config contracts.LLMConfig) Option {
	return func(a *Agent) { a.llmConfig = &config }
}

// WithCacheConfig sets prompt caching configuration.
func WithCacheConfig(config contracts.CacheConfig) Option {
	return func(a *Agent) { a.cacheConfig = &config }
}

// WithMaxIterations sets the maximum number of tool-calling iterations.
func WithMaxIterations(maxIterations int) Option {
	return func(a *Agent) { a.maxIterations = maxIterations }
}

// WithDisableFinalSummary disables the final summary LLM call.
func WithDisableFinalSummary(disable bool) Option {
	return func(a *Agent) { a.disableFinalSummary = disable }
}

// WithStreamConfig sets streaming configuration for the agent.
func WithStreamConfig(config *contracts.StreamConfig) Option {
	return func(a *Agent) { a.streamConfig = config }
}

// WithRemoteClient injects transport behavior for a remote agent.
func WithRemoteClient(url string, client contracts.RemoteAgentClient) Option {
	return func(a *Agent) {
		a.isRemote = true
		a.remoteURL = url
		a.remoteClient = client
		a.llm = nil
	}
}

// WithRemoteClientFactory enables transport-backed agent tools from configuration.
func WithRemoteClientFactory(factory func(string) contracts.RemoteAgentClient) Option {
	return func(a *Agent) { a.remoteClientFactory = factory }
}

// WithRemoteTimeout sets the timeout for remote agent operations.
func WithRemoteTimeout(timeout time.Duration) Option {
	return func(a *Agent) { a.remoteTimeout = timeout }
}

// WithCustomRunFunction replaces the default Run behavior.
func WithCustomRunFunction(fn CustomRunFunction) Option {
	return func(a *Agent) { a.customRunFunc = fn }
}

// WithCustomRunStreamFunction replaces the default RunStream behavior.
func WithCustomRunStreamFunction(fn CustomRunStreamFunction) Option {
	return func(a *Agent) { a.customRunStreamFunc = fn }
}
