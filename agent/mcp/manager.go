package mcp

import (
	"context"
	"fmt"
	"slices"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/config"
	"github.com/dm-vev/nu/internal/mcp/lazy"
)

// createLazyMCPTools creates lazy MCP tools from configurations
// LazyTools creates deferred tools from lazy MCP configurations.
func (m *Manager) LazyTools() []contracts.Tool {
	var lazyTools []contracts.Tool

	m.logger.Info(context.Background(), fmt.Sprintf("Creating lazy MCP tools from %d configs...", len(m.lazyConfigs)), nil)
	for _, cfg := range m.lazyConfigs {
		m.logger.Info(context.Background(), fmt.Sprintf("Processing MCP config: %s (type: %s)", cfg.Name, cfg.Type), nil)
		// Create lazy server config
		lazyServerConfig := config.Config{
			Name:              cfg.Name,
			Type:              cfg.Type,
			Command:           cfg.Command,
			Args:              cfg.Args,
			Env:               cfg.Env,
			URL:               cfg.URL,
			Token:             cfg.Token,
			HttpTransportMode: cfg.HttpTransportMode,
			AllowedTools:      cfg.AllowedTools,
		}

		// If no specific tools are defined, discover all tools from the server
		if len(cfg.Tools) == 0 {
			m.logger.Info(context.Background(), fmt.Sprintf("No tools specified for %s, discovering tools from server", cfg.Name), nil)

			// Create a temporary server instance to discover tools
			ctx := context.Background()
			server, err := lazy.GetOrCreateServerFromCache(ctx, lazyServerConfig)
			if err != nil {
				m.logger.Error(ctx, fmt.Sprintf("Failed to create server for tool discovery: %v", err), nil)
				continue
			}

			m.logger.Info(context.Background(), fmt.Sprintf("Discovered MCP server metadata for %s:", cfg.Name), nil)
			if serverInfo, err := server.GetServerInfo(); err == nil && serverInfo != nil {
				m.logger.Info(context.Background(), fmt.Sprintf("  Name: %s", serverInfo.Name), nil)
				if serverInfo.Title != "" {
					m.logger.Info(context.Background(), fmt.Sprintf("  Title: %s", serverInfo.Title), nil)
				}
				if serverInfo.Version != "" {
					m.logger.Info(context.Background(), fmt.Sprintf("  Version: %s", serverInfo.Version), nil)
				}
			}

			// Discover available tools from the server
			discoveredTools, err := server.ListTools(ctx)
			if err != nil {
				m.logger.Error(ctx, fmt.Sprintf("Failed to discover tools from %s: %v", cfg.Name, err), nil)
				continue
			}

			m.logger.Info(context.Background(), fmt.Sprintf("Discovered %d tools from %s server", len(discoveredTools), cfg.Name), nil)

			// Create lazy tools for each discovered tool
			for _, discoveredTool := range discoveredTools {
				if len(cfg.AllowedTools) > 0 && !slices.Contains(cfg.AllowedTools, discoveredTool.Name) {
					m.logger.Info(ctx, fmt.Sprintf("Skipping tool '%s'. Tool is not in allowed tools list - %q", discoveredTool.Name, cfg.AllowedTools), nil)
					continue
				}

				m.logger.Info(context.Background(), fmt.Sprintf("Creating lazy tool for %s: %s (Schema: %v)", discoveredTool.Name, discoveredTool.Description, discoveredTool.Schema), nil)

				lazyTool := lazy.NewLazyMCPTool(
					discoveredTool.Name,
					discoveredTool.Description,
					discoveredTool.Schema,
					lazyServerConfig,
				)
				lazyTools = append(lazyTools, lazyTool)
			}
		} else {
			// Create a temporary server instance to discover metadata even for configured tools
			ctx := context.Background()
			server, err := lazy.GetOrCreateServerFromCache(ctx, lazyServerConfig)
			if err != nil {
				m.logger.Warn(context.Background(), fmt.Sprintf("Failed to create server for metadata discovery: %v", err), nil)
			} else {
				// Log discovered server metadata
				if serverInfo, err := server.GetServerInfo(); err == nil && serverInfo != nil {
					m.logger.Info(context.Background(), fmt.Sprintf("Discovered MCP server metadata for %s: Name=%s, Title=%s, Version=%s",
						cfg.Name, serverInfo.Name, serverInfo.Title, serverInfo.Version), nil)
				}
			}

			// Create lazy tools for each configured tool
			for _, toolConfig := range cfg.Tools {
				m.logger.Info(context.Background(), fmt.Sprintf("Creating tool: %s", toolConfig.Name), nil)
				lazyTool := lazy.NewLazyMCPTool(
					toolConfig.Name,
					toolConfig.Description,
					toolConfig.Schema,
					lazyServerConfig,
				)
				lazyTools = append(lazyTools, lazyTool)
			}
		}
	}
	m.logger.Info(context.Background(), fmt.Sprintf("Created %d lazy MCP tools", len(lazyTools)), nil)
	return lazyTools
}
