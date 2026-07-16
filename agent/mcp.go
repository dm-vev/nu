package agent

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/agent/mcp"
	"github.com/dm-vev/nu/contracts"
)

// WithMCPServers sets the MCP servers for the agent.
func WithMCPServers(mcpServers []contracts.MCPServer) Option {
	return func(a *Agent) {
		a.mcpServers = mcpServers
	}
}

// WithLazyMCPConfigs sets the lazy MCP server configurations for the agent.
func WithLazyMCPConfigs(configs []LazyMCPConfig) Option {
	return func(a *Agent) {
		a.lazyMCPConfigs = configs
	}
}

// WithMCPURLs adds MCP servers from URL strings.
func WithMCPURLs(urls ...string) Option {
	return func(a *Agent) {
		configs, err := mcp.ConfigsFromURLs(urls...)
		if err != nil {
			a.logger.Warn(context.Background(), "Failed to parse some MCP URLs", map[string]interface{}{"error": err.Error()})
			return
		}
		a.lazyMCPConfigs = append(a.lazyMCPConfigs, configs...)
	}
}

// WithMCPPresets adds predefined MCP server configurations.
func WithMCPPresets(presetNames ...string) Option {
	return func(a *Agent) {
		configs, err := mcp.ConfigsFromPresets(presetNames...)
		if err != nil {
			a.logger.Warn(context.Background(), "Failed to load some MCP presets", map[string]interface{}{"error": err.Error()})
			return
		}
		a.lazyMCPConfigs = append(a.lazyMCPConfigs, configs...)
	}
}

func (a *Agent) initializeMCPTools() error {
	manager := mcp.NewManager(a.mcpServers, a.lazyMCPConfigs, a.logger)
	tools, err := manager.Initialize(context.Background())
	if err != nil {
		return err
	}
	a.tools = deduplicateTools(append(a.tools, tools...))
	return nil
}

func (a *Agent) collectMCPTools(ctx context.Context) ([]contracts.Tool, error) {
	return mcp.NewManager(a.mcpServers, a.lazyMCPConfigs, a.logger).CollectTools(ctx)
}

func (a *Agent) createLazyMCPTools() []contracts.Tool {
	return mcp.NewManager(a.mcpServers, a.lazyMCPConfigs, a.logger).LazyTools()
}

func (a *Agent) getAllToolsSync() []contracts.Tool {
	return a.tools
}

func applyMCPConfig(a *Agent, config *config.MCPConfiguration, configVars map[string]string) {
	a.lazyMCPConfigs = mcp.ApplyConfig(config, configVars, a.logger)
}

// GetMCPConfigFromAgent extracts MCP configuration from an agent.
func GetMCPConfigFromAgent(a *Agent) *config.MCPConfiguration {
	return mcp.ConfigurationFromLazy(a.lazyMCPConfigs)
}

// WithMCPConfigFromJSON adds MCP servers from a JSON configuration file.
func WithMCPConfigFromJSON(filePath string) Option {
	return func(a *Agent) {
		config, err := mcp.LoadJSON(filePath)
		if err != nil {
			return
		}
		fmt.Printf("MCP Config loaded from JSON: %s\n", filePath)
		applyMCPConfig(a, config, nil)
	}
}

// WithMCPConfigFromYAML adds MCP servers from a YAML configuration file.
func WithMCPConfigFromYAML(filePath string) Option {
	return func(a *Agent) {
		config, err := mcp.LoadYAML(filePath)
		if err != nil {
			return
		}
		applyMCPConfig(a, config, nil)
	}
}

// WithMCPConfig adds MCP servers from a configuration object.
func WithMCPConfig(config *config.MCPConfiguration) Option {
	return func(a *Agent) {
		applyMCPConfig(a, config, nil)
	}
}
