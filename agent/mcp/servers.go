package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/internal/mcp/builder"
	"github.com/dm-vev/nu/telemetry"
)

func buildLazyConfigs(servers map[string]config.MCPServerConfig, configVars map[string]string, mcpBuilder *builder.Builder, logger telemetry.Logger) []LazyMCPConfig {
	ctx := context.Background()
	lazyConfigs := make([]LazyMCPConfig, 0, len(servers))
	for serverName, serverConfig := range servers {
		serverType := serverConfig.GetServerType()

		switch serverType {
		case "stdio":
			mcpBuilder.AddStdioServer(serverName, serverConfig.Command, serverConfig.Args...)
			envSlice := make([]string, 0, len(serverConfig.Env))
			for key, value := range serverConfig.Env {
				resolvedValue := config.ExpandWithVariables(value, configVars)
				envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, resolvedValue))
			}
			lazyConfigs = append(lazyConfigs, LazyMCPConfig{
				Name:         serverName,
				Type:         "stdio",
				Command:      serverConfig.Command,
				Args:         serverConfig.Args,
				Env:          envSlice,
				Tools:        []LazyMCPToolConfig{},
				AllowedTools: serverConfig.AllowedTools,
			})

		case "http":
			if serverConfig.Token != "" {
				mcpBuilder.AddHTTPServerWithAuth(serverName, serverConfig.URL, serverConfig.Token)
			} else {
				mcpBuilder.AddHTTPServer(serverName, serverConfig.URL)
			}
			transportMode := strings.ToLower(serverConfig.HttpTransportMode)
			if transportMode == "" {
				transportMode = "sse"
			}
			lazyConfigs = append(lazyConfigs, LazyMCPConfig{
				Name:              serverName,
				Type:              "http",
				URL:               serverConfig.URL,
				Token:             serverConfig.Token,
				HttpTransportMode: transportMode,
				Tools:             []LazyMCPToolConfig{},
				AllowedTools:      serverConfig.AllowedTools,
			})

		default:
			if logger != nil {
				logger.Warn(ctx, "Unknown MCP server type", map[string]interface{}{"server_name": serverName, "server_type": serverType})
			}
			continue
		}

		if logger != nil {
			logger.Info(ctx, "Configured MCP server from config", map[string]interface{}{"server_name": serverName, "server_type": serverType})
		}
	}
	return lazyConfigs
}
