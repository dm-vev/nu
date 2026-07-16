package agent

import (
	"context"
	"time"

	agentconfig "nu/internal/agent/config"
	"nu/internal/contracts"
	agenttool "nu/internal/tools/agent"
)

// WithAgentConfig sets the agent configuration from a YAML config
func WithAgentConfig(config agentconfig.AgentConfig, variables map[string]string) Option {
	return func(a *Agent) {
		// Expand environment variables in all config sections
		expandedConfig := agentconfig.ExpandAgentConfig(config)

		// Existing system prompt processing
		systemPrompt := agentconfig.FormatSystemPromptFromConfig(expandedConfig, variables)
		a.systemPrompt = systemPrompt

		// Existing response format and MCP config
		if expandedConfig.ResponseFormat != nil {
			responseFormat, err := agentconfig.ConvertYAMLSchemaToResponseFormat(expandedConfig.ResponseFormat)
			if err == nil && responseFormat != nil {
				a.responseFormat = responseFormat
			}
		}

		// Extract configVars for MCP configuration expansion
		var configVars map[string]string
		if expandedConfig.ConfigSource != nil && expandedConfig.ConfigSource.Variables != nil {
			configVars = expandedConfig.ConfigSource.Variables
		} else {
			configVars = make(map[string]string)
		}

		if expandedConfig.MCP != nil {
			applyMCPConfig(a, expandedConfig.MCP, configVars)
		}

		// Apply behavioral settings
		if expandedConfig.MaxIterations != nil {
			a.maxIterations = *expandedConfig.MaxIterations
		}
		if expandedConfig.DisableFinalSummary != nil {
			a.disableFinalSummary = *expandedConfig.DisableFinalSummary
		}
		if expandedConfig.RequirePlanApproval != nil {
			a.requirePlanApproval = *expandedConfig.RequirePlanApproval
		}

		// Apply complex configuration objects
		if expandedConfig.StreamConfig != nil {
			a.streamConfig = agentconfig.ConvertStreamConfig(expandedConfig.StreamConfig)
		}
		if expandedConfig.LLMConfig != nil {
			a.llmConfig = agentconfig.ConvertLLMConfig(expandedConfig.LLMConfig)
		}
		if expandedConfig.CacheConfig != nil {
			a.cacheConfig = agentconfig.ConvertCacheConfig(expandedConfig.CacheConfig)
		}

		// Process LLM provider configuration
		if expandedConfig.LLMProvider != nil {
			if a.logger != nil {
				a.logger.Info(context.Background(), "Found LLM provider configuration in YAML", map[string]interface{}{
					"provider": expandedConfig.LLMProvider.Provider,
					"model":    expandedConfig.LLMProvider.Model,
					"has_llm":  a.llm != nil,
				})
			}
			if a.llm == nil {
				// Only create LLM from YAML if no LLM was provided programmatically
				llmClient, err := createLLMFromConfig(expandedConfig.LLMProvider)
				if err != nil {
					// Log warning but continue - don't fail agent creation for LLM issues
					if a.logger != nil {
						a.logger.Warn(context.Background(), "Failed to create LLM from YAML config", map[string]interface{}{
							"provider": expandedConfig.LLMProvider.Provider,
							"error":    err.Error(),
						})
					}
				} else {
					if a.logger != nil {
						a.logger.Info(context.Background(), "Successfully created LLM from YAML config", map[string]interface{}{
							"provider": expandedConfig.LLMProvider.Provider,
							"model":    expandedConfig.LLMProvider.Model,
						})
					}
					a.llm = llmClient
				}
			} else {
				if a.logger != nil {
					a.logger.Info(context.Background(), "LLM already provided programmatically, skipping YAML config", map[string]interface{}{
						"yaml_provider": expandedConfig.LLMProvider.Provider,
					})
				}
			}
		} else {
			if a.logger != nil {
				a.logger.Info(context.Background(), "No LLM provider configuration found in YAML", map[string]interface{}{
					"has_llm": a.llm != nil,
				})
			}
		}

		a.configuredTools = append(a.configuredTools, expandedConfig.Tools...)

		// Store memory config for later instantiation (after LLM is set)
		if expandedConfig.Memory != nil {
			a.memoryConfig = agentconfig.ConvertMemoryConfig(expandedConfig.Memory)
		}

		// Process image generation configuration
		if expandedConfig.ImageGeneration != nil {
			imgGenEnabled := expandedConfig.ImageGeneration.Enabled == nil || *expandedConfig.ImageGeneration.Enabled
			if imgGenEnabled {
				// Check if multi-turn editing is enabled
				multiTurnEnabled := false
				if expandedConfig.ImageGeneration.MultiTurnEditing != nil {
					multiTurnEnabled = expandedConfig.ImageGeneration.MultiTurnEditing.Enabled == nil || *expandedConfig.ImageGeneration.MultiTurnEditing.Enabled
				}

				// Create image generation tool (with multi-turn support if enabled)
				imgTool, err := createImageGenerationToolFromConfig(expandedConfig.ImageGeneration, a.logger)
				if err != nil {
					if a.logger != nil {
						a.logger.Warn(context.Background(), "Failed to create image generation tool from config", map[string]interface{}{
							"error": err.Error(),
						})
					}
				} else if imgTool != nil {
					a.tools = deduplicateTools(append(a.tools, imgTool))
					if a.logger != nil {
						a.logger.Info(context.Background(), "Successfully created image generation tool from YAML config", map[string]interface{}{
							"multi_turn_enabled": multiTurnEnabled,
						})
					}
				}
			}
		}

		// Apply runtime settings
		if expandedConfig.Runtime != nil {
			// TODO: Set log level if logger supports it when LogLevel is specified
			// Currently the logger interface doesn't support dynamic level setting
			if expandedConfig.Runtime.TimeoutDuration != "" {
				if timeout, err := time.ParseDuration(expandedConfig.Runtime.TimeoutDuration); err == nil {
					a.timeout = timeout
				}
			}
			if expandedConfig.Runtime.EnableTracing != nil && *expandedConfig.Runtime.EnableTracing {
				// Tracing enablement flag stored for later use
				a.tracingEnabled = true
			}
			if expandedConfig.Runtime.EnableMetrics != nil && *expandedConfig.Runtime.EnableMetrics {
				// Metrics enablement flag stored for later use
				a.metricsEnabled = true
			}
		}

		// Process sub-agents recursively
		if expandedConfig.SubAgents != nil {
			// Merge ConfigSource variables with OS env variables for sub-agents
			// ConfigSource variables take priority (they're from the config service)
			mergedVariables := make(map[string]string)
			// Start with OS env variables
			for k, v := range variables {
				mergedVariables[k] = v
			}
			// Override with ConfigSource variables if available
			if expandedConfig.ConfigSource != nil && expandedConfig.ConfigSource.Variables != nil {
				for k, v := range expandedConfig.ConfigSource.Variables {
					mergedVariables[k] = v
				}
			}

			subAgents, err := createSubAgentsFromConfig(expandedConfig.SubAgents, mergedVariables, a.llm, a.memory, a.tracer, a.logger)
			if err != nil {
				// Log error but don't fail agent creation
				if a.logger != nil {
					a.logger.Warn(context.Background(), "Failed to create some sub-agents from config", map[string]interface{}{
						"error": err.Error(),
					})
				}
			} else if len(subAgents) > 0 {
				// Add sub-agents using WithAgents
				a.subAgents = subAgents
				// Convert sub-agents to tools
				agentTools := make([]contracts.Tool, 0, len(subAgents))
				for _, subAgent := range subAgents {
					agentTool := agenttool.NewAgentTool(subAgent)
					// Pass logger and tracer if available on parent agent
					if a.logger != nil {
						agentTool = agentTool.WithLogger(a.logger)
					}
					if a.tracer != nil {
						agentTool = agentTool.WithTracer(a.tracer)
					}
					agentTools = append(agentTools, agentTool)
				}
				// Deduplicate before adding to agent
				a.tools = deduplicateTools(append(a.tools, agentTools...))
			}
		}

		// Store the expanded configuration for later access
		a.generatedAgentConfig = &expandedConfig
	}
}

func (a *Agent) initializeConfiguredTools() {
	if len(a.configuredTools) == 0 {
		return
	}
	factory := NewToolFactory(a.remoteClientFactory)
	for _, toolConfig := range a.configuredTools {
		if toolConfig.Enabled != nil && !*toolConfig.Enabled {
			continue
		}
		tool, err := factory.CreateTool(toolConfig)
		if err != nil {
			a.logger.Warn(context.Background(), "Failed to create tool from config", map[string]interface{}{
				"tool_name": toolConfig.Name,
				"tool_type": toolConfig.Type,
				"error":     err.Error(),
			})
			continue
		}
		a.tools = deduplicateTools(append(a.tools, tool))
	}
}
