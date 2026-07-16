package agent

import (
	"context"
	"fmt"
	"slices"

	"nu/internal/contracts"
	"nu/internal/mcp"
)

// WithMCPServers sets the MCP servers for the agent
func WithMCPServers(mcpServers []contracts.MCPServer) Option {
	return func(a *Agent) {
		a.mcpServers = mcpServers
	}
}

// WithLazyMCPConfigs sets the lazy MCP server configurations for the agent
func WithLazyMCPConfigs(configs []LazyMCPConfig) Option {
	return func(a *Agent) {
		a.lazyMCPConfigs = configs
	}
}

// WithMCPURLs adds MCP servers from URL strings
// Supports formats:
// - stdio://command/path/to/executable
// - http://localhost:8080/mcp
// - https://api.example.com/mcp?token=xxx
// - mcp://preset-name (for presets)
func WithMCPURLs(urls ...string) Option {
	return func(a *Agent) {
		builder := mcp.NewBuilder()
		for _, url := range urls {
			builder.AddServer(url)
		}

		// Build lazy configurations
		lazyConfigs, err := builder.BuildLazy()
		if err != nil {
			// Log error but don't fail agent creation
			if a.logger != nil {
				a.logger.Warn(context.Background(), "Failed to parse some MCP URLs", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return
		}

		// Convert mcp.LazyMCPServerConfig to agent.LazyMCPConfig
		for _, config := range lazyConfigs {
			agentConfig := LazyMCPConfig{
				Name:              config.Name,
				Type:              config.Type,
				Command:           config.Command,
				Args:              config.Args,
				Env:               config.Env,
				URL:               config.URL,
				Token:             config.Token,
				HttpTransportMode: config.HttpTransportMode,
				AllowedTools:      config.AllowedTools,
			}
			a.lazyMCPConfigs = append(a.lazyMCPConfigs, agentConfig)
		}
	}
}

// WithMCPPresets adds predefined MCP server configurations
func WithMCPPresets(presetNames ...string) Option {
	return func(a *Agent) {
		builder := mcp.NewBuilder()
		for _, preset := range presetNames {
			builder.AddPreset(preset)
		}

		// Build lazy configurations
		lazyConfigs, err := builder.BuildLazy()
		if err != nil {
			// Log error but don't fail agent creation
			if a.logger != nil {
				a.logger.Warn(context.Background(), "Failed to load some MCP presets", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return
		}

		// Convert mcp.LazyMCPServerConfig to agent.LazyMCPConfig
		for _, config := range lazyConfigs {
			agentConfig := LazyMCPConfig{
				Name:              config.Name,
				Type:              config.Type,
				Command:           config.Command,
				Args:              config.Args,
				Env:               config.Env,
				URL:               config.URL,
				Token:             config.Token,
				HttpTransportMode: config.HttpTransportMode,
				AllowedTools:      config.AllowedTools,
			}
			a.lazyMCPConfigs = append(a.lazyMCPConfigs, agentConfig)
		}
	}
}

// collectMCPTools collects tools from all MCP servers
func (a *Agent) collectMCPTools(ctx context.Context) ([]contracts.Tool, error) {
	var mcpTools []contracts.Tool

	for _, server := range a.mcpServers {
		// List tools from this server
		tools, err := server.ListTools(ctx)
		if err != nil {
			a.logger.Error(ctx, fmt.Sprintf("Failed to list tools from MCP server: %v", err), nil)
			continue
		}

		// Convert MCP tools to agent tools
		for _, mcpTool := range tools {
			// Create a new MCPTool
			tool := mcp.NewTool(mcpTool.Name, mcpTool.Description, mcpTool.Schema, server)
			mcpTools = append(mcpTools, tool)
		}
	}

	return mcpTools, nil
}

// createLazyMCPTools creates lazy MCP tools from configurations
func (a *Agent) createLazyMCPTools() []contracts.Tool {
	var lazyTools []contracts.Tool

	a.logger.Info(context.Background(), fmt.Sprintf("Creating lazy MCP tools from %d configs...", len(a.lazyMCPConfigs)), nil)
	for _, config := range a.lazyMCPConfigs {
		a.logger.Info(context.Background(), fmt.Sprintf("Processing MCP config: %s (type: %s)", config.Name, config.Type), nil)
		// Create lazy server config
		lazyServerConfig := mcp.LazyMCPServerConfig{
			Name:              config.Name,
			Type:              config.Type,
			Command:           config.Command,
			Args:              config.Args,
			Env:               config.Env,
			URL:               config.URL,
			Token:             config.Token,
			HttpTransportMode: config.HttpTransportMode,
			AllowedTools:      config.AllowedTools,
		}

		// If no specific tools are defined, discover all tools from the server
		if len(config.Tools) == 0 {
			a.logger.Info(context.Background(), fmt.Sprintf("No tools specified for %s, discovering tools from server", config.Name), nil)

			// Create a temporary server instance to discover tools
			ctx := context.Background()
			server, err := mcp.GetOrCreateServerFromCache(ctx, lazyServerConfig)
			if err != nil {
				a.logger.Error(ctx, fmt.Sprintf("Failed to create server for tool discovery: %v", err), nil)
				continue
			}

			a.logger.Info(context.Background(), fmt.Sprintf("Discovered MCP server metadata for %s:", config.Name), nil)
			if serverInfo, err := server.GetServerInfo(); err == nil && serverInfo != nil {
				a.logger.Info(context.Background(), fmt.Sprintf("  Name: %s", serverInfo.Name), nil)
				if serverInfo.Title != "" {
					a.logger.Info(context.Background(), fmt.Sprintf("  Title: %s", serverInfo.Title), nil)
				}
				if serverInfo.Version != "" {
					a.logger.Info(context.Background(), fmt.Sprintf("  Version: %s", serverInfo.Version), nil)
				}
			}

			// Discover available tools from the server
			discoveredTools, err := server.ListTools(ctx)
			if err != nil {
				a.logger.Error(ctx, fmt.Sprintf("Failed to discover tools from %s: %v", config.Name, err), nil)
				continue
			}

			a.logger.Info(context.Background(), fmt.Sprintf("Discovered %d tools from %s server", len(discoveredTools), config.Name), nil)

			// Create lazy tools for each discovered tool
			for _, discoveredTool := range discoveredTools {
				if len(config.AllowedTools) > 0 && !slices.Contains(config.AllowedTools, discoveredTool.Name) {
					a.logger.Info(ctx, fmt.Sprintf("Skipping tool '%s'. Tool is not in allowed tools list - %q", discoveredTool.Name, config.AllowedTools), nil)
					continue
				}

				a.logger.Info(context.Background(), fmt.Sprintf("Creating lazy tool for %s: %s (Schema: %v)", discoveredTool.Name, discoveredTool.Description, discoveredTool.Schema), nil)

				lazyTool := mcp.NewLazyMCPTool(
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
			server, err := mcp.GetOrCreateServerFromCache(ctx, lazyServerConfig)
			if err != nil {
				a.logger.Warn(context.Background(), fmt.Sprintf("Failed to create server for metadata discovery: %v", err), nil)
			} else {
				// Log discovered server metadata
				if serverInfo, err := server.GetServerInfo(); err == nil && serverInfo != nil {
					a.logger.Info(context.Background(), fmt.Sprintf("Discovered MCP server metadata for %s: Name=%s, Title=%s, Version=%s",
						config.Name, serverInfo.Name, serverInfo.Title, serverInfo.Version), nil)
				}
			}

			// Create lazy tools for each configured tool
			for _, toolConfig := range config.Tools {
				a.logger.Info(context.Background(), fmt.Sprintf("Creating tool: %s", toolConfig.Name), nil)
				lazyTool := mcp.NewLazyMCPTool(
					toolConfig.Name,
					toolConfig.Description,
					toolConfig.Schema,
					lazyServerConfig,
				)
				lazyTools = append(lazyTools, lazyTool)
			}
		}
	}
	a.logger.Info(context.Background(), fmt.Sprintf("Created %d lazy MCP tools", len(lazyTools)), nil)
	return lazyTools
}

// initializeMCPTools eagerly initializes MCP tools during agent creation
func (a *Agent) initializeMCPTools() error {
	ctx := context.Background()

	// Initialize regular MCP tools if available
	if len(a.mcpServers) > 0 {
		mcpTools, err := a.collectMCPTools(ctx)
		if err != nil {
			return fmt.Errorf("failed to collect MCP server tools: %w", err)
		}
		// Add MCP tools to the main tools slice with deduplication
		a.tools = deduplicateTools(append(a.tools, mcpTools...))
		a.logger.Info(context.Background(), fmt.Sprintf("Initialized %d MCP server tools", len(mcpTools)), nil)
	}

	// Initialize lazy MCP tools if available
	if len(a.lazyMCPConfigs) > 0 {
		lazyMCPTools := a.createLazyMCPTools()
		// Add lazy MCP tools to the main tools slice with deduplication
		a.tools = deduplicateTools(append(a.tools, lazyMCPTools...))
		a.logger.Info(context.Background(), fmt.Sprintf("Initialized %d lazy MCP tools", len(lazyMCPTools)), nil)
	}

	return nil
}

// getAllToolsSync returns all tools (manual + MCP) synchronously for use during initialization
func (a *Agent) getAllToolsSync() []contracts.Tool {
	// At this point, a.tools already contains manual tools + initialized MCP tools
	return a.tools
}
