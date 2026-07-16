package mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nu/internal/contracts"
	"nu/internal/telemetry"
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
func NewHTTPWithRetry(ctx context.Context, config HTTPConfig, retryConfig *RetryConfig) (contracts.MCPServer, error) {
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

func NewCustomTransportServerWithRetry(ctx context.Context, config CustomTransportServerConfig, retryConfig *RetryConfig) (contracts.MCPServer, error) {
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

func newServerFromTransport(ctx context.Context, transport mcp.Transport, serverName, serverType string, retryConfig *RetryConfig, logger telemetry.Logger) (contracts.MCPServer, *Error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "agent-sdk-go", Version: "0.0.0"}, nil)
	client.AddSendingMiddleware(tracingMiddleware)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, ClassifyError(err, "Connect", serverName, serverType)
	}

	initResult := session.InitializeResult()
	var serverInfo *contracts.MCPServerInfo
	var capabilities *contracts.MCPServerCapabilities
	if initResult != nil {
		if initResult.ServerInfo != nil {
			serverInfo = &contracts.MCPServerInfo{
				Name: initResult.ServerInfo.Name, Title: initResult.ServerInfo.Title,
				Version: initResult.ServerInfo.Version,
			}
			logger.Info(ctx, "Discovered MCP server metadata", map[string]interface{}{
				"server_name": serverInfo.Name, "server_title": serverInfo.Title,
				"server_version": serverInfo.Version,
			})
		}
		if initResult.Capabilities != nil {
			capabilities = convertMCPCapabilities(initResult.Capabilities)
		}
	}
	logger.Debug(ctx, "MCP server connection established with metadata - ", map[string]interface{}{
		"has_server_info": serverInfo != nil, "has_capabilities": capabilities != nil,
	})
	server := &Server{
		session: session, logger: logger, serverInfo: serverInfo, capabilities: capabilities,
	}
	if retryConfig != nil {
		return NewRetryableServer(server, retryConfig), nil
	}
	return server, nil
}
