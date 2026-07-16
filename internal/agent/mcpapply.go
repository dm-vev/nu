package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	agentconfig "nu/internal/agent/config"
	"nu/internal/mcp/builder"
)

// applyMCPConfig applies MCP configuration to an agent
// configVars contains variables from ConfigSource (config service) for expansion
func applyMCPConfig(a *Agent, config *agentconfig.MCPConfiguration, configVars map[string]string) {
	if config == nil {
		return
	}

	ctx := context.Background()

	// Create MCP builder
	mcpBuilder := builder.NewBuilder()

	// Apply global configuration if present, with defaults
	globalConfig := config.Global
	if globalConfig == nil {
		// Set defaults when no global config provided
		trueVal := true
		globalConfig = &agentconfig.MCPGlobalConfig{
			Timeout:         "30s",
			RetryAttempts:   3,
			HealthCheck:     &trueVal,
			EnableResources: &trueVal,
			EnablePrompts:   &trueVal,
			EnableSampling:  &trueVal,
			EnableSchemas:   &trueVal,
			LogLevel:        "info",
		}
	} else {
		// Apply defaults for unspecified values
		if globalConfig.Timeout == "" {
			globalConfig.Timeout = "30s"
		}
		if globalConfig.RetryAttempts == 0 {
			globalConfig.RetryAttempts = 3
		}
		// Set defaults for nil pointers (unspecified values)
		if globalConfig.HealthCheck == nil {
			trueVal := true
			globalConfig.HealthCheck = &trueVal
		}
		if globalConfig.EnableResources == nil {
			trueVal := true
			globalConfig.EnableResources = &trueVal
		}
		if globalConfig.EnablePrompts == nil {
			trueVal := true
			globalConfig.EnablePrompts = &trueVal
		}
		if globalConfig.EnableSampling == nil {
			trueVal := true
			globalConfig.EnableSampling = &trueVal
		}
		if globalConfig.EnableSchemas == nil {
			trueVal := true
			globalConfig.EnableSchemas = &trueVal
		}
		if globalConfig.LogLevel == "" {
			globalConfig.LogLevel = "info"
		}
	}

	// Apply timeout to builder
	if globalConfig.Timeout != "" {
		if timeout, err := time.ParseDuration(globalConfig.Timeout); err == nil {
			mcpBuilder.WithTimeout(timeout)
			if a.logger != nil {
				a.logger.Debug(ctx, "MCP timeout configured", map[string]interface{}{
					"timeout": globalConfig.Timeout,
				})
			}
		}
	}

	// Apply retry attempts to builder
	if globalConfig.RetryAttempts > 0 {
		mcpBuilder.WithRetry(globalConfig.RetryAttempts, 1*time.Second)
		if a.logger != nil {
			a.logger.Debug(ctx, "MCP retry attempts configured", map[string]interface{}{
				"retry_attempts": globalConfig.RetryAttempts,
			})
		}
	}

	// Apply health check to builder
	mcpBuilder.WithHealthCheck(*globalConfig.HealthCheck)

	if a.logger != nil {
		a.logger.Debug(ctx, "MCP global configuration applied", map[string]interface{}{
			"health_check":     *globalConfig.HealthCheck,
			"enable_resources": *globalConfig.EnableResources,
			"enable_prompts":   *globalConfig.EnablePrompts,
			"enable_sampling":  *globalConfig.EnableSampling,
			"enable_schemas":   *globalConfig.EnableSchemas,
			"log_level":        globalConfig.LogLevel,
			"timeout":          globalConfig.Timeout,
			"retry_attempts":   globalConfig.RetryAttempts,
		})
	}

	// Convert server configurations to lazy MCP configs
	var lazyConfigs []LazyMCPConfig
	enabledCount := 0

	for serverName, serverConfig := range config.MCPServers {
		serverType := serverConfig.GetServerType()

		switch serverType {
		case "stdio":
			mcpBuilder.AddStdioServer(serverName, serverConfig.Command, serverConfig.Args...)

			// Convert environment map to string slice format
			// Resolve environment variable placeholders using configVars first, then OS env
			var envSlice []string
			for key, value := range serverConfig.Env {
				// Use expandWithConfigVars to check ConfigSource variables first, then OS env
				resolvedValue := agentconfig.ExpandWithVariables(value, configVars)
				envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, resolvedValue))
			}

			lazyConfig := LazyMCPConfig{
				Name:         serverName,
				Type:         "stdio",
				Command:      serverConfig.Command,
				Args:         serverConfig.Args,
				Env:          envSlice,
				Tools:        []LazyMCPToolConfig{}, // Will discover dynamically
				AllowedTools: serverConfig.AllowedTools,
			}
			lazyConfigs = append(lazyConfigs, lazyConfig)

		case "http":
			if serverConfig.Token != "" {
				mcpBuilder.AddHTTPServerWithAuth(serverName, serverConfig.URL, serverConfig.Token)
			} else {
				mcpBuilder.AddHTTPServer(serverName, serverConfig.URL)
			}

			lazyConfig := LazyMCPConfig{
				Name:         serverName,
				Type:         "http",
				URL:          serverConfig.URL,
				Token:        serverConfig.Token,    // Preserve token for lazy initialization
				Tools:        []LazyMCPToolConfig{}, // Will discover dynamically
				AllowedTools: serverConfig.AllowedTools,
			}
			if serverConfig.HttpTransportMode != "" {
				// handle case-insensitivity
				lazyConfig.HttpTransportMode = strings.ToLower(serverConfig.HttpTransportMode)
			} else {
				lazyConfig.HttpTransportMode = "sse" // Default to sse
			}
			lazyConfigs = append(lazyConfigs, lazyConfig)

		default:
			if a.logger != nil {
				a.logger.Warn(ctx, "Unknown MCP server type", map[string]interface{}{
					"server_name": serverName,
					"server_type": serverType,
				})
			}
			continue
		}

		enabledCount++
		if a.logger != nil {
			a.logger.Info(ctx, "Configured MCP server from config", map[string]interface{}{
				"server_name": serverName,
				"server_type": serverType,
			})
		}
	}

	// Set lazy MCP configs on agent
	a.lazyMCPConfigs = lazyConfigs

	if a.logger != nil {
		a.logger.Info(ctx, "Applied MCP configuration", map[string]interface{}{
			"total_servers":   len(config.MCPServers),
			"enabled_servers": enabledCount,
		})
	}
}
