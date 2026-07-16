package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// Server is the implementation of contracts.MCPServer using the official SDK
type Server struct {
	session      *mcp.ClientSession
	logger       telemetry.Logger
	serverInfo   *contracts.MCPServerInfo
	capabilities *contracts.MCPServerCapabilities
}

// convertMCPCapabilities converts mcp.ServerCapabilities to contracts.MCPServerCapabilities
func convertMCPCapabilities(caps *mcp.ServerCapabilities) *contracts.MCPServerCapabilities {
	if caps == nil {
		return nil
	}

	result := &contracts.MCPServerCapabilities{}
	if caps.Tools != nil {
		result.Tools = &contracts.MCPToolCapabilities{ListChanged: caps.Tools.ListChanged}
	}
	if caps.Resources != nil {
		result.Resources = &contracts.MCPResourceCapabilities{
			Subscribe:   caps.Resources.Subscribe,
			ListChanged: caps.Resources.ListChanged,
		}
	}
	if caps.Prompts != nil {
		result.Prompts = &contracts.MCPPromptCapabilities{ListChanged: caps.Prompts.ListChanged}
	}
	return result
}

// NewClient creates a new MCPServer with the given transport using the official SDK
func NewClient(ctx context.Context, transport mcp.Transport) (contracts.MCPServer, error) {
	logger := telemetry.NewLogger()
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "agent-sdk-go",
		Version: "0.0.0",
	}, nil)
	client.AddSendingMiddleware(tracingMiddleware)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		logger.Error(ctx, "Failed to connect to MCP server", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	initResult := session.InitializeResult()
	var serverInfo *contracts.MCPServerInfo
	var capabilities *contracts.MCPServerCapabilities
	if initResult != nil {
		if initResult.ServerInfo != nil {
			serverInfo = &contracts.MCPServerInfo{
				Name:    initResult.ServerInfo.Name,
				Title:   initResult.ServerInfo.Title,
				Version: initResult.ServerInfo.Version,
			}
			logger.Info(ctx, "Discovered MCP server metadata", map[string]interface{}{
				"server_name":    serverInfo.Name,
				"server_title":   serverInfo.Title,
				"server_version": serverInfo.Version,
			})
		}
		if initResult.Capabilities != nil {
			capabilities = convertMCPCapabilities(initResult.Capabilities)
		}
	}

	logger.Debug(ctx, "MCP server connection established with metadata", map[string]interface{}{
		"has_server_info":  serverInfo != nil,
		"has_capabilities": capabilities != nil,
	})
	return &Server{
		session:      session,
		logger:       logger,
		serverInfo:   serverInfo,
		capabilities: capabilities,
	}, nil
}

// Initialize initializes the connection to the MCP server
func (s *Server) Initialize(ctx context.Context) error {
	return nil
}

// GetServerInfo returns the server metadata discovered during initialization
func (s *Server) GetServerInfo() (*contracts.MCPServerInfo, error) {
	return s.serverInfo, nil
}

// GetCapabilities returns the server capabilities discovered during initialization
func (s *Server) GetCapabilities() (*contracts.MCPServerCapabilities, error) {
	return s.capabilities, nil
}

// Close closes the connection to the MCP server
func (s *Server) Close() error {
	s.logger.Debug(context.Background(), "Closing MCP server connection", nil)
	err := s.session.Close()
	if err != nil {
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(context.Background(), "Failed to close MCP server connection", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		s.logger.Debug(context.Background(), "MCP server connection closed successfully", nil)
	}
	return err
}
