package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/fault"
)

// ListResources lists the resources available on the MCP server
func (s *Server) ListResources(ctx context.Context) ([]contracts.MCPResource, error) {
	s.logger.Debug(ctx, "Listing MCP resources", nil)
	resp, err := s.session.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		mcpErr := fault.ClassifyError(err, "ListResources", "server", "unknown")
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(ctx, "Failed to list MCP resources", map[string]interface{}{
			"error":      err.Error(),
			"error_type": mcpErr.ErrorType,
			"retryable":  mcpErr.Retryable,
		})
		return nil, mcpErr
	}

	resources := make([]contracts.MCPResource, 0, len(resp.Resources))
	for _, r := range resp.Resources {
		resource := contracts.MCPResource{
			URI:         r.URI,
			Name:        r.Name,
			Description: r.Description,
			MimeType:    r.MIMEType,
			Metadata:    make(map[string]string),
		}
		if r.Annotations != nil {
			if len(r.Annotations.Audience) > 0 {
				var audienceStrs []string
				for _, role := range r.Annotations.Audience {
					audienceStrs = append(audienceStrs, string(role))
				}
				resource.Metadata["audience"] = strings.Join(audienceStrs, ",")
			}
			if r.Annotations.LastModified != "" {
				resource.Metadata["lastModified"] = r.Annotations.LastModified
			}
			if r.Annotations.Priority > 0 {
				resource.Metadata["priority"] = fmt.Sprintf("%.2f", r.Annotations.Priority)
			}
		}
		resources = append(resources, resource)
	}
	s.logger.Info(ctx, "Successfully listed MCP resources", map[string]interface{}{
		"resource_count": len(resources),
	})
	return resources, nil
}

// GetResource retrieves a specific resource by URI
func (s *Server) GetResource(ctx context.Context, uri string) (*contracts.MCPResourceContent, error) {
	s.logger.Debug(ctx, "Getting MCP resource", map[string]interface{}{"uri": uri})
	resp, err := s.session.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	if err != nil {
		mcpErr := fault.ClassifyError(err, "GetResource", "server", "unknown")
		mcpErr = mcpErr.WithMetadata("uri", uri)
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(ctx, "Failed to get MCP resource", map[string]interface{}{
			"uri":        uri,
			"error":      err.Error(),
			"error_type": mcpErr.ErrorType,
			"retryable":  mcpErr.Retryable,
		})
		return nil, mcpErr
	}

	content := &contracts.MCPResourceContent{URI: uri, Metadata: make(map[string]string)}
	if len(resp.Contents) > 0 {
		firstContent := resp.Contents[0]
		if firstContent.Text != "" {
			content.Text = firstContent.Text
			content.MimeType = firstContent.MIMEType
			if content.MimeType == "" {
				content.MimeType = "text/plain"
			}
		} else if len(firstContent.Blob) > 0 {
			content.Blob = firstContent.Blob
			content.MimeType = firstContent.MIMEType
			if content.MimeType == "" {
				content.MimeType = "application/octet-stream"
			}
		}
	}
	s.logger.Debug(ctx, "Successfully retrieved MCP resource", map[string]interface{}{
		"uri":       uri,
		"mime_type": content.MimeType,
		"size":      len(content.Text) + len(content.Blob),
	})
	return content, nil
}

// WatchResource watches for changes to a resource (if supported)
func (s *Server) WatchResource(ctx context.Context, uri string) (<-chan contracts.MCPResourceUpdate, error) {
	s.logger.Debug(ctx, "Setting up resource watch", map[string]interface{}{"uri": uri})
	updates := make(chan contracts.MCPResourceUpdate, 10)
	go func() {
		defer close(updates)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		var lastContent *contracts.MCPResourceContent
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				content, err := s.GetResource(ctx, uri)
				if err != nil {
					updates <- contracts.MCPResourceUpdate{
						URI: uri, Type: contracts.MCPResourceUpdateTypeError,
						Timestamp: time.Now(), Error: err,
					}
					continue
				}
				if lastContent == nil || content.Text != lastContent.Text || !bytes.Equal(content.Blob, lastContent.Blob) {
					updates <- contracts.MCPResourceUpdate{
						URI: uri, Type: contracts.MCPResourceUpdateTypeChanged,
						Content: content, Timestamp: time.Now(),
					}
					lastContent = content
				}
			}
		}
	}()
	return updates, nil
}
