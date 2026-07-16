package mcp

import (
	"context"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/internal/mcp/builder"
	"github.com/dm-vev/nu/telemetry"
)

// ApplyConfig applies MCP configuration and returns lazy server configurations.
// configVars contains variables from ConfigSource for expansion.
func ApplyConfig(config *config.MCPConfiguration, configVars map[string]string, logger telemetry.Logger) []LazyMCPConfig {
	if config == nil {
		return nil
	}

	ctx := context.Background()
	mcpBuilder := builder.NewBuilder()
	globalConfig := applyGlobalDefaults(config.Global)
	configureBuilder(ctx, mcpBuilder, globalConfig, logger)
	lazyConfigs := buildLazyConfigs(config.MCPServers, configVars, mcpBuilder, logger)

	if logger != nil {
		logger.Info(ctx, "Applied MCP configuration", map[string]interface{}{
			"total_servers":   len(config.MCPServers),
			"enabled_servers": len(lazyConfigs),
		})
	}
	return lazyConfigs
}
