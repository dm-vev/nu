package config

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nu/internal/telemetry"
)

// Config describes an MCP server that can be created on demand.
type Config struct {
	Name                string
	Type                string
	Command             string
	Args                []string
	Env                 []string
	URL                 string
	Token               string
	HttpTransportMode   string
	AllowedTools        []string
	CustomMCPTransport  mcp.Transport
	Logger              telemetry.Logger
	CustomTransportType string
}
