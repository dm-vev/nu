package mcp

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// RegistryManager integrates registry discovery with MCP server creation
type RegistryManager struct {
	registryClient *RegistryClient
	builder        *Builder
	logger         telemetry.Logger
}

func NewRegistryManager(registryURL string) *RegistryManager {
	return &RegistryManager{
		registryClient: NewRegistryClient(registryURL),
		builder:        NewBuilder(), logger: telemetry.NewLogger(),
	}
}

func (rm *RegistryManager) DiscoverAndInstallServer(ctx context.Context, serverID string) (*LazyMCPServerConfig, error) {
	server, err := rm.registryClient.GetServer(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover server: %w", err)
	}
	config, err := rm.registryServerToConfig(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create server config: %w", err)
	}
	rm.logger.Info(ctx, "Successfully discovered and configured server from registry", map[string]interface{}{
		"server_id": serverID, "name": server.Name, "type": config.Type,
	})
	return config, nil
}

func (rm *RegistryManager) DiscoverServersByCapability(ctx context.Context, capability string) ([]*RegistryServer, error) {
	response, err := rm.registryClient.SearchServers(ctx, capability)
	if err != nil {
		return nil, err
	}
	var matchingServers []*RegistryServer
	for i, server := range response.Servers {
		if rm.serverProvidesCapability(&response.Servers[i], capability) {
			matchingServers = append(matchingServers, &server)
		}
	}
	return matchingServers, nil
}

func (rm *RegistryManager) BuildServerFromRegistry(ctx context.Context, serverID string) (contracts.MCPServer, error) {
	config, err := rm.DiscoverAndInstallServer(ctx, serverID)
	if err != nil {
		return nil, err
	}
	server, err := rm.builder.initializeServer(ctx, *config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}
	return server, nil
}

func (rm *RegistryManager) registryServerToConfig(server *RegistryServer) (*LazyMCPServerConfig, error) {
	config := &LazyMCPServerConfig{Name: server.Name}
	switch server.Installation.Type {
	case "stdio":
		config.Type = "stdio"
		config.Command = server.Installation.Command
		config.Args = server.Installation.Args
		if len(server.Installation.Env) > 0 {
			for key, value := range server.Installation.Env {
				config.Env = append(config.Env, fmt.Sprintf("%s=%s", key, value))
			}
		}
	case "npm":
		config.Type = "stdio"
		config.Command = "npx"
		config.Args = []string{server.Installation.Command}
		if len(server.Installation.Args) > 0 {
			config.Args = append(config.Args, server.Installation.Args...)
		}
	case "pip", "python":
		config.Type = "stdio"
		config.Command = "python"
		config.Args = []string{"-m", server.Installation.Command}
		if len(server.Installation.Args) > 0 {
			config.Args = append(config.Args, server.Installation.Args...)
		}
	case "docker":
		config.Type = "stdio"
		config.Command = "docker"
		config.Args = []string{"run", "--rm", "-i", server.Installation.Command}
		if len(server.Installation.Args) > 0 {
			config.Args = append(config.Args, server.Installation.Args...)
		}
	case "http":
		config.Type = "http"
		config.URL = server.Installation.Command
	default:
		return nil, fmt.Errorf("unsupported installation type: %s", server.Installation.Type)
	}
	return config, nil
}

func (rm *RegistryManager) serverProvidesCapability(server *RegistryServer, capability string) bool {
	capability = strings.ToLower(capability)
	for _, tool := range server.Tools {
		if strings.Contains(strings.ToLower(tool.Name), capability) ||
			strings.Contains(strings.ToLower(tool.Description), capability) ||
			strings.Contains(strings.ToLower(tool.Category), capability) {
			return true
		}
	}
	for _, resource := range server.Resources {
		if strings.Contains(strings.ToLower(resource.Type), capability) ||
			strings.Contains(strings.ToLower(resource.Description), capability) {
			return true
		}
	}
	for _, prompt := range server.Prompts {
		if strings.Contains(strings.ToLower(prompt.Name), capability) ||
			strings.Contains(strings.ToLower(prompt.Description), capability) ||
			strings.Contains(strings.ToLower(prompt.Category), capability) {
			return true
		}
	}
	for _, tag := range server.Tags {
		if strings.Contains(strings.ToLower(tag), capability) {
			return true
		}
	}
	return strings.Contains(strings.ToLower(server.Description), capability) ||
		strings.Contains(strings.ToLower(server.Category), capability)
}

func (rm *RegistryManager) DiscoverFileSystemServers(ctx context.Context) ([]*RegistryServer, error) {
	return rm.DiscoverServersByCapability(ctx, "filesystem")
}

func (rm *RegistryManager) DiscoverDatabaseServers(ctx context.Context) ([]*RegistryServer, error) {
	return rm.DiscoverServersByCapability(ctx, "database")
}

func (rm *RegistryManager) DiscoverWebServers(ctx context.Context) ([]*RegistryServer, error) {
	return rm.DiscoverServersByCapability(ctx, "web")
}

func (rm *RegistryManager) DiscoverCodeServers(ctx context.Context) ([]*RegistryServer, error) {
	return rm.DiscoverServersByCapability(ctx, "code")
}

func (rm *RegistryManager) GetPopularServers(ctx context.Context) (*SearchResponse, error) {
	return rm.registryClient.ListServers(ctx, &SearchOptions{Verified: true, Limit: 20})
}
