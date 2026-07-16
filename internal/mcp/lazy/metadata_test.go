package lazy

import (
	"testing"

	"nu/internal/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test lazy MCP tool with server metadata
func TestLazyMCPToolWithMetadata(t *testing.T) {
	// Create a lazy MCP server config
	config := LazyMCPServerConfig{
		Name:    "test-server",
		Type:    "stdio",
		Command: "test-command",
	}

	// Create a lazy MCP tool
	tool := NewLazyMCPTool("test-tool", "Test tool description", nil, config)
	lazyTool, ok := tool.(*LazyMCPTool)
	require.True(t, ok)

	// Test initial state
	assert.Equal(t, "test-tool", lazyTool.Name())
	assert.Equal(t, "Test tool description", lazyTool.Description())

	// Test fallback description when tool description is empty
	emptyDescTool := NewLazyMCPTool("test-tool-2", "", nil, config)
	lazyEmptyTool, ok := emptyDescTool.(*LazyMCPTool)
	require.True(t, ok)

	// Initially should use generic description
	assert.Equal(t, "test-tool-2 (MCP tool)", lazyEmptyTool.Description())

	// Simulate server metadata being loaded
	lazyEmptyTool.serverInfo = &contracts.MCPServerInfo{
		Name:  "test-server",
		Title: "Test Server",
	}

	// Now description should include server context
	assert.Equal(t, "test-tool-2 (from Test Server)", lazyEmptyTool.Description())
}

// Test server metadata cache functionality
func TestServerMetadataCache(t *testing.T) {
	config := LazyMCPServerConfig{
		Name:    "cache-test-server",
		Type:    "stdio",
		Command: "test-command",
	}

	// Test getting metadata from cache when none exists
	metadata := GetServerMetadataFromCache(config)
	assert.Nil(t, metadata)

	// Simulate storing metadata in cache
	serverKey := "stdio:cache-test-server:test-command"
	globalServerCache.mu.Lock()
	globalServerCache.serverMetadata[serverKey] = &contracts.MCPServerInfo{
		Name:    "cache-test-server",
		Title:   "Cached Test Server",
		Version: "v2.0.0",
	}
	globalServerCache.mu.Unlock()

	// Test getting metadata from cache
	metadata = GetServerMetadataFromCache(config)
	require.NotNil(t, metadata)
	assert.Equal(t, "cache-test-server", metadata.Name)
	assert.Equal(t, "Cached Test Server", metadata.Title)
	assert.Equal(t, "v2.0.0", metadata.Version)

	// Clean up
	globalServerCache.mu.Lock()
	delete(globalServerCache.serverMetadata, serverKey)
	globalServerCache.mu.Unlock()
}
