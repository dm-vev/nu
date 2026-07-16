package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/client"
	"github.com/dm-vev/nu/internal/mcp/fault"
	"github.com/dm-vev/nu/internal/mcp/retry"
	"github.com/dm-vev/nu/telemetry"
)

// HTTPConfig holds configuration for an HTTP MCP server
type HTTPConfig struct {
	BaseURL      string
	Path         string
	Token        string
	ProtocolType ServerProtocolType
	Logger       telemetry.Logger

	ResourceIndicator string `json:"resource_indicator,omitempty"`
}

// CustomTransportServerConfig holds configuration for a custom transport MCP server
type CustomTransportServerConfig struct {
	Transport     mcp.Transport
	Logger        telemetry.Logger
	TransportType string
}

// ServerProtocolType defines the protocol type for the MCP server communication
type ServerProtocolType string

const (
	StreamableHTTP ServerProtocolType = "streamable"
	SSE            ServerProtocolType = "sse"
)

type customRoundTripper struct {
	delegate http.RoundTripper
	token    string
}

func (rt *customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+rt.token)
	return rt.delegate.RoundTrip(req)
}

func customHTTPClient(token string) *http.Client {
	return &http.Client{Transport: &customRoundTripper{
		delegate: http.DefaultTransport,
		token:    token,
	}}
}

// NewServer creates a new MCPServer that communicates over HTTP using the official SDK
func NewServer(ctx context.Context, config HTTPConfig) (contracts.MCPServer, error) {
	return NewHTTPWithRetry(ctx, config, nil)
}

// NewHTTPWithRetry creates a new HTTP MCPServer with retry logic
func NewHTTPWithRetry(ctx context.Context, config HTTPConfig, retryConfig *retry.Config) (contracts.MCPServer, error) {
	httpClient := http.DefaultClient
	if config.Token != "" {
		httpClient = customHTTPClient(config.Token)
	}

	var transport mcp.Transport
	switch config.ProtocolType {
	case SSE:
		transport = &mcp.SSEClientTransport{Endpoint: config.BaseURL, HTTPClient: httpClient}
	case StreamableHTTP:
		transport = &mcp.StreamableClientTransport{Endpoint: config.BaseURL, HTTPClient: httpClient}
	default:
		config.Logger.Warn(ctx, "Server protocol type is not set, defaulting to SSE", map[string]interface{}{})
		transport = &mcp.SSEClientTransport{Endpoint: config.BaseURL, HTTPClient: httpClient}
	}

	server, err := newServerFromTransport(ctx, transport, "http-server", "http", retryConfig, config.Logger)
	if err != nil {
		config.Logger.Error(ctx, "[HTTP SERVER ERROR] Failed to connect to MCP server", map[string]interface{}{
			"error": err.Error(), "error_type": err.ErrorType, "retryable": err.Retryable,
		})
		return nil, err
	}
	return server, nil
}

func NewCustomTransportServer(ctx context.Context, config CustomTransportServerConfig) (contracts.MCPServer, error) {
	return NewCustomTransportServerWithRetry(ctx, config, nil)
}

func NewCustomTransportServerWithRetry(ctx context.Context, config CustomTransportServerConfig, retryConfig *retry.Config) (contracts.MCPServer, error) {
	serverName := strings.ToLower(config.TransportType) + "-server"
	server, err := newServerFromTransport(ctx, config.Transport, serverName, config.TransportType, retryConfig, config.Logger)
	if err != nil {
		config.Logger.Error(ctx, "[SERVER ERROR] Failed to connect to MCP server - ", map[string]interface{}{
			"error": err.Error(), "error_type": err.ErrorType, "retryable": err.Retryable,
			"transport_type": config.TransportType, "server_name": serverName,
		})
		return nil, err
	}
	return server, nil
}

func newServerFromTransport(ctx context.Context, transport mcp.Transport, serverName, serverType string, retryConfig *retry.Config, logger telemetry.Logger) (contracts.MCPServer, *fault.Error) {
	server, err := client.NewClientWithLogger(ctx, transport, logger)
	if err != nil {
		return nil, fault.ClassifyError(err, "Connect", serverName, serverType)
	}
	if retryConfig != nil {
		return retry.New(server, retryConfig), nil
	}
	return server, nil
}
