package resource

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceManager(t *testing.T) {
	server1 := &testkit.Server{}
	server2 := &testkit.Server{}
	servers := []contracts.MCPServer{server1, server2}

	manager := NewResourceManager(servers)

	assert.NotNil(t, manager)
	assert.Len(t, manager.servers, 2)
	assert.NotNil(t, manager.logger)
	assert.Equal(t, servers, manager.servers)
}

func TestResourceManager_ListAllResources(t *testing.T) {
	ctx := context.Background()

	resources1 := []contracts.MCPResource{
		{URI: "file://test1.txt", Name: "test1", Description: "First test file"},
		{URI: "file://test2.txt", Name: "test2", Description: "Second test file"},
	}

	resources2 := []contracts.MCPResource{
		{URI: "file://test3.txt", Name: "test3", Description: "Third test file"},
	}

	t.Run("successful listing from all servers", func(t *testing.T) {
		server1 := &testkit.Server{}
		server2 := &testkit.Server{}
		servers := []contracts.MCPServer{server1, server2}

		server1.On("ListResources", ctx).Return(resources1, nil)
		server2.On("ListResources", ctx).Return(resources2, nil)

		manager := NewResourceManager(servers)
		result, err := manager.ListAllResources(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, resources1, result["server-0"])
		assert.Equal(t, resources2, result["server-1"])

		server1.AssertExpectations(t)
		server2.AssertExpectations(t)
	})

	t.Run("one server fails, continues with others", func(t *testing.T) {
		server1 := &testkit.Server{}
		server2 := &testkit.Server{}
		servers := []contracts.MCPServer{server1, server2}

		server1.On("ListResources", ctx).Return(nil, errors.New("connection failed"))
		server2.On("ListResources", ctx).Return(resources2, nil)

		manager := NewResourceManager(servers)
		result, err := manager.ListAllResources(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, resources2, result["server-1"])
		assert.NotContains(t, result, "server-0")

		server1.AssertExpectations(t)
		server2.AssertExpectations(t)
	})

	t.Run("empty servers list", func(t *testing.T) {
		manager := NewResourceManager([]contracts.MCPServer{})
		result, err := manager.ListAllResources(ctx)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestResourceManager_FindResources(t *testing.T) {
	ctx := context.Background()

	resources := []contracts.MCPResource{
		{
			URI:         "file://code.go",
			Name:        "Go Code File",
			Description: "Go source code",
			MimeType:    "text/x-go",
		},
		{
			URI:         "file://doc.md",
			Name:        "Documentation",
			Description: "Project documentation",
			MimeType:    "text/markdown",
		},
		{
			URI:         "file://image.png",
			Name:        "Image File",
			Description: "PNG image for documentation",
			MimeType:    "image/png",
		},
	}

	tests := []struct {
		name            string
		pattern         string
		expectedMatches int
		expectedNames   []string
	}{
		{
			name:            "search by name",
			pattern:         "code",
			expectedMatches: 1,
			expectedNames:   []string{"Go Code File"},
		},
		{
			name:            "search by description",
			pattern:         "documentation",
			expectedMatches: 2, // Both doc.md and image.png have "documentation" in description
			expectedNames:   []string{"Documentation", "Image File"},
		},
		{
			name:            "search by file extension",
			pattern:         ".go",
			expectedMatches: 1,
			expectedNames:   []string{"Go Code File"},
		},
		{
			name:            "search by URI",
			pattern:         "file://",
			expectedMatches: 3, // All files have this URI prefix
			expectedNames:   []string{"Go Code File", "Documentation", "Image File"},
		},
		{
			name:            "case insensitive search",
			pattern:         "PNG",
			expectedMatches: 1,
			expectedNames:   []string{"Image File"},
		},
		{
			name:            "no matches",
			pattern:         "nonexistent",
			expectedMatches: 0,
			expectedNames:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &testkit.Server{}
			server.On("ListResources", ctx).Return(resources, nil)

			manager := NewResourceManager([]contracts.MCPServer{server})
			matches, err := manager.FindResources(ctx, tt.pattern)

			assert.NoError(t, err)
			assert.Len(t, matches, tt.expectedMatches)

			matchedNames := make([]string, len(matches))
			for i, match := range matches {
				matchedNames[i] = match.Resource.Name
			}
			assert.ElementsMatch(t, tt.expectedNames, matchedNames)

			server.AssertExpectations(t)
		})
	}
}

func TestResourceManager_GetResourceContent(t *testing.T) {
	ctx := context.Background()
	uri := "file://test.txt"

	expectedContent := &contracts.MCPResourceContent{
		URI:      uri,
		MimeType: "text/plain",
		Text:     "File content",
	}

	t.Run("successful resource retrieval", func(t *testing.T) {
		server1 := &testkit.Server{}
		server2 := &testkit.Server{}
		servers := []contracts.MCPServer{server1, server2}

		server1.On("GetResource", ctx, uri).Return(nil, errors.New("not found"))
		server2.On("GetResource", ctx, uri).Return(expectedContent, nil)

		manager := NewResourceManager(servers)
		content, err := manager.GetResourceContent(ctx, uri)

		assert.NoError(t, err)
		require.NotNil(t, content)
		assert.Equal(t, expectedContent, content)

		server1.AssertExpectations(t)
		server2.AssertExpectations(t)
	})

	t.Run("resource not found on any server", func(t *testing.T) {
		server := &testkit.Server{}
		server.On("GetResource", ctx, uri).Return(nil, errors.New("not found"))

		manager := NewResourceManager([]contracts.MCPServer{server})
		content, err := manager.GetResourceContent(ctx, uri)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found on any server")
		assert.Nil(t, content)

		server.AssertExpectations(t)
	})

	t.Run("first server succeeds", func(t *testing.T) {
		server1 := &testkit.Server{}
		server2 := &testkit.Server{}
		servers := []contracts.MCPServer{server1, server2}

		server1.On("GetResource", ctx, uri).Return(expectedContent, nil)
		// server2 should not be called since server1 succeeded

		manager := NewResourceManager(servers)
		content, err := manager.GetResourceContent(ctx, uri)

		assert.NoError(t, err)
		require.NotNil(t, content)
		assert.Equal(t, expectedContent, content)

		server1.AssertExpectations(t)
		server2.AssertNotCalled(t, "GetResource")
	})
}

func TestResourceManager_WatchResources(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uris := []string{"file://test1.txt", "file://test2.txt"}

	t.Run("successful watch setup", func(t *testing.T) {
		server := &testkit.Server{}
		servers := []contracts.MCPServer{server}

		// Create channels for server updates
		updates1 := make(chan contracts.MCPResourceUpdate, 1)
		updates2 := make(chan contracts.MCPResourceUpdate, 1)

		server.On("WatchResource", ctx, uris[0]).Return((<-chan contracts.MCPResourceUpdate)(updates1), nil)
		server.On("WatchResource", ctx, uris[1]).Return((<-chan contracts.MCPResourceUpdate)(updates2), nil)

		manager := NewResourceManager(servers)
		combinedUpdates, err := manager.WatchResources(ctx, uris)

		assert.NoError(t, err)
		assert.NotNil(t, combinedUpdates)

		// Send test updates
		testUpdate1 := contracts.MCPResourceUpdate{URI: uris[0], Type: contracts.MCPResourceUpdateType("modified")}
		testUpdate2 := contracts.MCPResourceUpdate{URI: uris[1], Type: contracts.MCPResourceUpdateType("deleted")}

		updates1 <- testUpdate1
		updates2 <- testUpdate2

		// Read from combined channel (order is non-deterministic)
		update1 := <-combinedUpdates
		update2 := <-combinedUpdates

		// Collect both updates
		receivedUpdates := []ResourceUpdate{update1, update2}

		// Verify both servers are "server-0"
		assert.Equal(t, "server-0", update1.Server)
		assert.Equal(t, "server-0", update2.Server)

		// Verify we received both expected updates (order doesn't matter)
		foundUpdate1 := false
		foundUpdate2 := false
		for _, u := range receivedUpdates {
			if u.Update.URI == testUpdate1.URI && u.Update.Type == testUpdate1.Type {
				foundUpdate1 = true
			}
			if u.Update.URI == testUpdate2.URI && u.Update.Type == testUpdate2.Type {
				foundUpdate2 = true
			}
		}
		assert.True(t, foundUpdate1, "Expected to find testUpdate1")
		assert.True(t, foundUpdate2, "Expected to find testUpdate2")

		server.AssertExpectations(t)
	})

	t.Run("server watch failure", func(t *testing.T) {
		server := &testkit.Server{}
		servers := []contracts.MCPServer{server}

		server.On("WatchResource", ctx, uris[0]).Return(nil, errors.New("watch failed"))

		manager := NewResourceManager(servers)
		combinedUpdates, err := manager.WatchResources(ctx, uris[:1])

		assert.NoError(t, err)
		assert.NotNil(t, combinedUpdates)

		server.AssertExpectations(t)
	})
}

func TestResourceManager_GetResourcesByType(t *testing.T) {
	ctx := context.Background()

	resources := []contracts.MCPResource{
		{URI: "file://text.txt", MimeType: "text/plain"},
		{URI: "file://code.go", MimeType: "text/x-go"},
		{URI: "file://image.png", MimeType: "image/png"},
		{URI: "file://image.jpg", MimeType: "image/jpeg"},
		{URI: "file://data.json", MimeType: "application/json"},
		{URI: "file://unknown", MimeType: ""}, // No MIME type
	}

	tests := []struct {
		name         string
		mimeType     string
		expectedURIs []string
	}{
		{
			name:         "exact MIME type match",
			mimeType:     "text/plain",
			expectedURIs: []string{"file://text.txt"},
		},
		{
			name:         "wildcard image type",
			mimeType:     "image/*",
			expectedURIs: []string{"file://image.png", "file://image.jpg"},
		},
		{
			name:         "wildcard text type",
			mimeType:     "text/*",
			expectedURIs: []string{"file://text.txt", "file://code.go"},
		},
		{
			name:         "no matches",
			mimeType:     "video/mp4",
			expectedURIs: []string{},
		},
		{
			name:         "specific application type",
			mimeType:     "application/json",
			expectedURIs: []string{"file://data.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &testkit.Server{}
			server.On("ListResources", ctx).Return(resources, nil)

			manager := NewResourceManager([]contracts.MCPServer{server})
			matches, err := manager.GetResourcesByType(ctx, tt.mimeType)

			assert.NoError(t, err)

			matchedURIs := make([]string, len(matches))
			for i, match := range matches {
				matchedURIs[i] = match.Resource.URI
			}
			assert.ElementsMatch(t, tt.expectedURIs, matchedURIs)

			server.AssertExpectations(t)
		})
	}
}

func TestResourceManager_matchesPattern(t *testing.T) {
	manager := NewResourceManager(nil)

	resource := contracts.MCPResource{
		URI:         "file://MyProject/Code.go",
		Name:        "Go Source File",
		Description: "Main application code",
		MimeType:    "text/x-go",
	}

	tests := []struct {
		pattern  string
		expected bool
	}{
		{"code", true},        // Name match (case insensitive)
		{"CODE", true},        // Name match (case insensitive)
		{"myproject", true},   // URI match
		{"application", true}, // Description match
		{".go", true},         // File extension match
		{"file://", true},     // URI prefix match
		{"nonexistent", false},
		{"", true}, // Empty pattern matches everything due to Contains
	}

	for _, tt := range tests {
		t.Run("pattern_"+tt.pattern, func(t *testing.T) {
			result := manager.matchesPattern(resource, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceManager_matchesMimeType(t *testing.T) {
	manager := NewResourceManager(nil)

	tests := []struct {
		name       string
		resource   contracts.MCPResource
		targetType string
		expected   bool
	}{
		{
			name:       "exact match",
			resource:   contracts.MCPResource{MimeType: "text/plain"},
			targetType: "text/plain",
			expected:   true,
		},
		{
			name:       "wildcard match",
			resource:   contracts.MCPResource{MimeType: "image/png"},
			targetType: "image/*",
			expected:   true,
		},
		{
			name:       "wildcard no match",
			resource:   contracts.MCPResource{MimeType: "text/plain"},
			targetType: "image/*",
			expected:   false,
		},
		{
			name:       "no mime type",
			resource:   contracts.MCPResource{MimeType: ""},
			targetType: "text/plain",
			expected:   false,
		},
		{
			name:       "complex mime type with params",
			resource:   contracts.MCPResource{MimeType: "text/html; charset=utf-8"},
			targetType: "text/html",
			expected:   true,
		},
		{
			name:       "invalid target mime type",
			resource:   contracts.MCPResource{MimeType: "text/plain"},
			targetType: "invalid/mime/type/format",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.matchesMimeType(tt.resource, tt.targetType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTextResource(t *testing.T) {
	tests := []struct {
		name     string
		resource contracts.MCPResource
		expected bool
	}{
		{
			name:     "no mime type defaults to text",
			resource: contracts.MCPResource{MimeType: ""},
			expected: true,
		},
		{
			name:     "plain text",
			resource: contracts.MCPResource{MimeType: "text/plain"},
			expected: true,
		},
		{
			name:     "html text",
			resource: contracts.MCPResource{MimeType: "text/html"},
			expected: true,
		},
		{
			name:     "application json",
			resource: contracts.MCPResource{MimeType: "application/json"},
			expected: true,
		},
		{
			name:     "application xml",
			resource: contracts.MCPResource{MimeType: "application/xml"},
			expected: true,
		},
		{
			name:     "custom +json",
			resource: contracts.MCPResource{MimeType: "application/vnd.api+json"},
			expected: true,
		},
		{
			name:     "custom +xml",
			resource: contracts.MCPResource{MimeType: "application/soap+xml"},
			expected: true,
		},
		{
			name:     "binary image",
			resource: contracts.MCPResource{MimeType: "image/png"},
			expected: false,
		},
		{
			name:     "binary video",
			resource: contracts.MCPResource{MimeType: "video/mp4"},
			expected: false,
		},
		{
			name:     "binary application",
			resource: contracts.MCPResource{MimeType: "application/octet-stream"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTextResource(tt.resource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBinaryResource(t *testing.T) {
	tests := []struct {
		name     string
		resource contracts.MCPResource
		expected bool
	}{
		{
			name:     "text file is not binary",
			resource: contracts.MCPResource{MimeType: "text/plain"},
			expected: false,
		},
		{
			name:     "image is binary",
			resource: contracts.MCPResource{MimeType: "image/png"},
			expected: true,
		},
		{
			name:     "no mime type defaults to text (not binary)",
			resource: contracts.MCPResource{MimeType: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBinaryResource(tt.resource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetResourceExtension(t *testing.T) {
	tests := []struct {
		name     string
		resource contracts.MCPResource
		expected string
	}{
		{
			name:     "file with extension",
			resource: contracts.MCPResource{URI: "file://path/test.txt"},
			expected: ".txt",
		},
		{
			name:     "file with uppercase extension",
			resource: contracts.MCPResource{URI: "file://path/TEST.TXT"},
			expected: ".txt", // Should be lowercase
		},
		{
			name:     "file with multiple dots",
			resource: contracts.MCPResource{URI: "file://path/test.tar.gz"},
			expected: ".gz", // Only last extension
		},
		{
			name:     "file without extension",
			resource: contracts.MCPResource{URI: "file://path/README"},
			expected: "",
		},
		{
			name:     "URI without file path",
			resource: contracts.MCPResource{URI: "http://example.com/api"},
			expected: "",
		},
		{
			name:     "empty URI",
			resource: contracts.MCPResource{URI: ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetResourceExtension(tt.resource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkResourceManager_FindResources(b *testing.B) {
	ctx := context.Background()
	resources := make([]contracts.MCPResource, 100)
	for i := 0; i < 100; i++ {
		resources[i] = contracts.MCPResource{
			URI:         fmt.Sprintf("file://test_%d.txt", i),
			Name:        fmt.Sprintf("Test File %d", i),
			Description: fmt.Sprintf("Test file number %d", i),
			MimeType:    "text/plain",
		}
	}

	server := &testkit.Server{}
	server.On("ListResources", ctx).Return(resources, nil)

	manager := NewResourceManager([]contracts.MCPServer{server})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.FindResources(ctx, "test")
		if err != nil {
			assert.Fail(b, "FindResources failed", err)
		}
	}
}

func BenchmarkResourceManager_matchesPattern(b *testing.B) {
	manager := NewResourceManager(nil)
	resource := contracts.MCPResource{
		URI:         "file://very/long/path/to/some/deep/directory/structure/test_file.go",
		Name:        "A Very Long Resource Name With Many Words",
		Description: "This is a very long description that contains many words and phrases that might be searched for",
		MimeType:    "text/x-go",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.matchesPattern(resource, "test")
	}
}

func BenchmarkIsTextResource(b *testing.B) {
	resource := contracts.MCPResource{MimeType: "text/html; charset=utf-8"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsTextResource(resource)
	}
}

func BenchmarkGetResourceExtension(b *testing.B) {
	resource := contracts.MCPResource{URI: "file://very/long/path/to/some/file.with.multiple.extensions.tar.gz"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetResourceExtension(resource)
	}
}
