package lazy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/transport"
	"github.com/dm-vev/nu/telemetry"
)

const (
	defaultMaxRetryAttempts = 5
	defaultRetryInterval    = 3 * time.Second
)

// LazyMCPServerCache manages shared MCP server instances
type LazyMCPServerCache struct {
	servers        map[string]contracts.MCPServer
	serverMetadata map[string]*contracts.MCPServerInfo
	mu             sync.RWMutex
	logger         telemetry.Logger
}

var globalServerCache = &LazyMCPServerCache{
	servers:        make(map[string]contracts.MCPServer),
	serverMetadata: make(map[string]*contracts.MCPServerInfo),
	logger:         telemetry.NewLogger(),
}

func (cache *LazyMCPServerCache) getOrCreateServer(ctx context.Context, config LazyMCPServerConfig) (contracts.MCPServer, error) {
	serverKey := fmt.Sprintf("%s:%s:%v:%s", config.Type, config.Name, config.Command, config.CustomTransportType)
	cache.mu.RLock()
	if server, exists := cache.servers[serverKey]; exists {
		cache.mu.RUnlock()
		return server, nil
	}
	cache.mu.RUnlock()

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if server, exists := cache.servers[serverKey]; exists {
		return server, nil
	}

	serverLogger := config.Logger
	if serverLogger == nil {
		serverLogger = telemetry.NewLogger()
	}
	var server contracts.MCPServer
	var err error
	switch config.Type {
	case "stdio":
		envDebug := make(map[string]string)
		for _, envVar := range config.Env {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				key, value := parts[0], parts[1]
				if len(value) > 10 {
					envDebug[key] = value[:10] + "..."
				} else if value == "" {
					envDebug[key] = "<EMPTY>"
				} else {
					envDebug[key] = value
				}
			}
		}
		cache.logger.Info(ctx, "Initializing MCP server on demand", map[string]interface{}{
			"server_name": config.Name, "server_type": config.Type,
			"command": config.Command, "args": config.Args,
			"env_count": len(config.Env), "env_vars": envDebug,
		})
		server, err = transport.NewStdioServer(ctx, transport.StdioServerConfig{
			Command: config.Command, Args: config.Args, Env: config.Env, Logger: serverLogger,
		})
	case "http":
		cache.logger.Info(ctx, "Initializing MCP server on demand", map[string]interface{}{
			"server_name": config.Name, "server_type": config.Type,
			"transport_mode": config.HttpTransportMode,
		})
		server, err = transport.NewServer(ctx, transport.HTTPConfig{
			BaseURL: config.URL, Token: config.Token,
			ProtocolType: transport.ServerProtocolType(config.HttpTransportMode), Logger: serverLogger,
		})
	case "custom":
		if config.CustomMCPTransport == nil {
			return nil, fmt.Errorf("custom MCP transport is required for 'custom' server type")
		}
		if config.CustomTransportType == "" {
			return nil, fmt.Errorf("custom transport type is required for 'custom' server type")
		}
		cache.logger.Info(ctx, "Initializing MCP server on demand", map[string]interface{}{
			"server_name": config.Name, "server_type": config.Type,
			"custom_transport_type": config.CustomTransportType,
		})
		server, err = transport.NewCustomTransportServer(ctx, transport.CustomTransportServerConfig{
			Transport: config.CustomMCPTransport, Logger: serverLogger,
			TransportType: config.CustomTransportType,
		})
	default:
		return nil, fmt.Errorf("unsupported MCP server type: %s", config.Type)
	}
	if err != nil {
		cache.logger.Error(ctx, "Failed to initialize MCP server", map[string]interface{}{
			"server_name": config.Name, "error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize MCP server '%s': %v", config.Name, err)
	}

	cache.servers[serverKey] = server
	if serverInfo, err := server.GetServerInfo(); err == nil && serverInfo != nil {
		cache.serverMetadata[serverKey] = serverInfo
		cache.logger.Info(ctx, "MCP server initialized successfully with metadata", map[string]interface{}{
			"server_name": config.Name, "discovered_name": serverInfo.Name,
			"discovered_title": serverInfo.Title, "discovered_version": serverInfo.Version,
		})
	} else {
		cache.logger.Info(ctx, "MCP server initialized successfully", map[string]interface{}{
			"server_name": config.Name,
		})
	}

	cache.logger.Info(ctx, "Waiting for MCP server to be ready", map[string]interface{}{
		"server_name": config.Name, "max_retries": defaultMaxRetryAttempts,
		"retry_interval": defaultRetryInterval.String(),
	})
	for attempt := 1; attempt <= defaultMaxRetryAttempts; attempt++ {
		_, err := server.ListTools(ctx)
		if err == nil {
			cache.logger.Info(ctx, "MCP server is ready", map[string]interface{}{
				"server_name": config.Name, "attempt": attempt,
			})
			break
		}
		if attempt < defaultMaxRetryAttempts {
			cache.logger.Debug(ctx, "MCP server not ready, retrying", map[string]interface{}{
				"server_name": config.Name, "attempt": attempt, "error": err.Error(),
			})
			time.Sleep(defaultRetryInterval)
		} else {
			cache.logger.Warn(ctx, "MCP server may not be fully ready after retries", map[string]interface{}{
				"server_name": config.Name, "attempts": attempt, "last_error": err.Error(),
			})
		}
	}
	return server, nil
}

// GetOrCreateServerFromCache provides public access to the global server cache
func GetOrCreateServerFromCache(ctx context.Context, config LazyMCPServerConfig) (contracts.MCPServer, error) {
	return globalServerCache.getOrCreateServer(ctx, config)
}

// GetServerMetadataFromCache gets server metadata from the global cache
func GetServerMetadataFromCache(config LazyMCPServerConfig) *contracts.MCPServerInfo {
	serverKey := fmt.Sprintf("%s:%s:%v", config.Type, config.Name, config.Command)
	globalServerCache.mu.RLock()
	defer globalServerCache.mu.RUnlock()
	return globalServerCache.serverMetadata[serverKey]
}
