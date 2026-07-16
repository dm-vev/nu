package mcp

import (
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// LazyMCPConfig describes an MCP server initialized on first tool use.
type LazyMCPConfig struct {
	Name              string
	Type              string
	Command           string
	Args              []string
	Env               []string
	URL               string
	Token             string
	Tools             []LazyMCPToolConfig
	HttpTransportMode string
	AllowedTools      []string
}

// LazyMCPToolConfig describes one lazy MCP tool.
type LazyMCPToolConfig struct {
	Name        string
	Description string
	Schema      interface{}
}

// Manager owns MCP server discovery and lazy tool initialization for an agent.
type Manager struct {
	servers     []contracts.MCPServer
	lazyConfigs []LazyMCPConfig
	logger      telemetry.Logger
}

// NewManager creates an MCP manager for one agent.
func NewManager(servers []contracts.MCPServer, lazyConfigs []LazyMCPConfig, logger telemetry.Logger) *Manager {
	return &Manager{servers: servers, lazyConfigs: lazyConfigs, logger: logger}
}
