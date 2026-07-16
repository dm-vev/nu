package agent

import (
	"context"
	"time"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/agent/image"
	"github.com/dm-vev/nu/agent/mcp"
	"github.com/dm-vev/nu/agent/providers"
	"github.com/dm-vev/nu/contracts"
	agenttool "github.com/dm-vev/nu/internal/tools/agent"
	"github.com/dm-vev/nu/telemetry"
)

type configResult struct {
	ExpandedConfig  config.AgentConfig
	SystemPrompt    string
	ResponseFormat  *contracts.ResponseFormat
	MCPConfigs      []mcp.LazyMCPConfig
	LLM             contracts.LLM
	ConfiguredTools []config.ToolConfigYAML
	MemoryConfig    map[string]interface{}
	ImageTool       contracts.Tool
	ImageToolError  error
	Timeout         time.Duration
	TracingEnabled  bool
	MetricsEnabled  bool
}

// WithAgentConfig sets the agent configuration from a YAML config.
func WithAgentConfig(cfg config.AgentConfig, variables map[string]string) Option {
	return func(a *Agent) {
		resolved := resolveConfig(cfg, variables, a.llm, a.logger)
		a.systemPrompt = resolved.SystemPrompt
		a.responseFormat = resolved.ResponseFormat
		a.lazyMCPConfigs = resolved.MCPConfigs
		a.llm = resolved.LLM
		a.configuredTools = append(a.configuredTools, resolved.ConfiguredTools...)
		a.memoryConfig = resolved.MemoryConfig
		if resolved.ImageToolError != nil && a.logger != nil {
			a.logger.Warn(context.Background(), "Failed to create image generation tool from config", map[string]interface{}{
				"error": resolved.ImageToolError.Error(),
			})
		}
		if resolved.ImageTool != nil {
			a.tools = deduplicateTools(append(a.tools, resolved.ImageTool))
		}
		if resolved.Timeout > 0 {
			a.timeout = resolved.Timeout
		}
		if resolved.TracingEnabled {
			a.tracingEnabled = true
		}
		if resolved.MetricsEnabled {
			a.metricsEnabled = true
		}

		expandedConfig := resolved.ExpandedConfig
		if expandedConfig.MaxIterations != nil {
			a.maxIterations = *expandedConfig.MaxIterations
		}
		if expandedConfig.DisableFinalSummary != nil {
			a.disableFinalSummary = *expandedConfig.DisableFinalSummary
		}
		if expandedConfig.RequirePlanApproval != nil {
			a.requirePlanApproval = *expandedConfig.RequirePlanApproval
		}
		if expandedConfig.StreamConfig != nil {
			a.streamConfig = config.ConvertStreamConfig(expandedConfig.StreamConfig)
		}
		if expandedConfig.LLMConfig != nil {
			a.llmConfig = config.ConvertLLMConfig(expandedConfig.LLMConfig)
		}
		if expandedConfig.CacheConfig != nil {
			a.cacheConfig = config.ConvertCacheConfig(expandedConfig.CacheConfig)
		}

		if expandedConfig.SubAgents != nil {
			mergedVariables := make(map[string]string, len(variables))
			for key, value := range variables {
				mergedVariables[key] = value
			}
			if expandedConfig.ConfigSource != nil {
				for key, value := range expandedConfig.ConfigSource.Variables {
					mergedVariables[key] = value
				}
			}

			subAgents, err := createSubAgentsFromConfig(expandedConfig.SubAgents, mergedVariables, a.llm, a.memory, a.tracer, a.logger)
			if err != nil {
				if a.logger != nil {
					a.logger.Warn(context.Background(), "Failed to create some sub-agents from config", map[string]interface{}{
						"error": err.Error(),
					})
				}
			} else if len(subAgents) > 0 {
				a.subAgents = subAgents
				agentTools := make([]contracts.Tool, 0, len(subAgents))
				for _, subAgent := range subAgents {
					agentTool := agenttool.NewAgentTool(subAgent)
					if a.logger != nil {
						agentTool = agentTool.WithLogger(a.logger)
					}
					if a.tracer != nil {
						agentTool = agentTool.WithTracer(a.tracer)
					}
					agentTools = append(agentTools, agentTool)
				}
				a.tools = deduplicateTools(append(a.tools, agentTools...))
			}
		}

		a.generatedAgentConfig = &expandedConfig
	}
}

func resolveConfig(cfg config.AgentConfig, variables map[string]string, currentLLM contracts.LLM, logger telemetry.Logger) configResult {
	expandedConfig := config.ExpandAgentConfig(cfg)
	result := configResult{
		ExpandedConfig:  expandedConfig,
		SystemPrompt:    config.FormatSystemPromptFromConfig(expandedConfig, variables),
		LLM:             currentLLM,
		ConfiguredTools: expandedConfig.Tools,
	}

	if expandedConfig.ResponseFormat != nil {
		responseFormat, err := config.ConvertYAMLSchemaToResponseFormat(expandedConfig.ResponseFormat)
		if err == nil && responseFormat != nil {
			result.ResponseFormat = responseFormat
		}
	}

	configVars := map[string]string{}
	if expandedConfig.ConfigSource != nil && expandedConfig.ConfigSource.Variables != nil {
		configVars = expandedConfig.ConfigSource.Variables
	}
	if expandedConfig.MCP != nil {
		result.MCPConfigs = mcp.ApplyConfig(expandedConfig.MCP, configVars, logger)
	}

	if expandedConfig.MaxIterations != nil {
		result.ExpandedConfig.MaxIterations = expandedConfig.MaxIterations
	}
	if expandedConfig.Memory != nil {
		result.MemoryConfig = config.ConvertMemoryConfig(expandedConfig.Memory)
	}

	if expandedConfig.LLMProvider != nil {
		if logger != nil {
			logger.Info(context.Background(), "Found LLM provider configuration in YAML", map[string]interface{}{
				"provider": expandedConfig.LLMProvider.Provider,
				"model":    expandedConfig.LLMProvider.Model,
				"has_llm":  currentLLM != nil,
			})
		}
		if currentLLM == nil {
			llmClient, err := providers.CreateLLMFromConfig(expandedConfig.LLMProvider)
			if err != nil {
				if logger != nil {
					logger.Warn(context.Background(), "Failed to create LLM from YAML config", map[string]interface{}{
						"provider": expandedConfig.LLMProvider.Provider,
						"error":    err.Error(),
					})
				}
			} else {
				result.LLM = llmClient
				if logger != nil {
					logger.Info(context.Background(), "Successfully created LLM from YAML config", map[string]interface{}{
						"provider": expandedConfig.LLMProvider.Provider,
						"model":    expandedConfig.LLMProvider.Model,
					})
				}
			}
		} else if logger != nil {
			logger.Info(context.Background(), "LLM already provided programmatically, skipping YAML config", map[string]interface{}{
				"yaml_provider": expandedConfig.LLMProvider.Provider,
			})
		}
	} else if logger != nil {
		logger.Info(context.Background(), "No LLM provider configuration found in YAML", map[string]interface{}{
			"has_llm": result.LLM != nil,
		})
	}

	if expandedConfig.ImageGeneration != nil {
		enabled := expandedConfig.ImageGeneration.Enabled == nil || *expandedConfig.ImageGeneration.Enabled
		if enabled {
			result.ImageTool, result.ImageToolError = image.CreateGenerationTool(expandedConfig.ImageGeneration, logger)
		}
	}

	if expandedConfig.Runtime != nil {
		if expandedConfig.Runtime.TimeoutDuration != "" {
			if timeout, err := time.ParseDuration(expandedConfig.Runtime.TimeoutDuration); err == nil {
				result.Timeout = timeout
			}
		}
		result.TracingEnabled = expandedConfig.Runtime.EnableTracing != nil && *expandedConfig.Runtime.EnableTracing
		result.MetricsEnabled = expandedConfig.Runtime.EnableMetrics != nil && *expandedConfig.Runtime.EnableMetrics
	}

	return result
}
