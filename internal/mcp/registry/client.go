package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"nu/internal/telemetry"
)

// RegistryClient handles interaction with MCP Registry for server discovery
type RegistryClient struct {
	baseURL    string
	httpClient *http.Client
	logger     telemetry.Logger
}

const DefaultRegistryURL = "https://registry.modelcontextprotocol.io"

func NewRegistryClient(baseURL string) *RegistryClient {
	if baseURL == "" {
		baseURL = DefaultRegistryURL
	}
	return &RegistryClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     telemetry.NewLogger(),
	}
}

func (rc *RegistryClient) ListServers(ctx context.Context, opts *SearchOptions) (*SearchResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Query != "" {
			params.Set("q", opts.Query)
		}
		if opts.Category != "" {
			params.Set("category", opts.Category)
		}
		if opts.Author != "" {
			params.Set("author", opts.Author)
		}
		if opts.Verified {
			params.Set("verified", "true")
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if len(opts.Tags) > 0 {
			params.Set("tags", strings.Join(opts.Tags, ","))
		}
	}
	fullURL := rc.baseURL + "/api/v1/servers"
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "agent-sdk-go/1.0")
	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Failed to close response body: %v\n", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry request failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}
	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	rc.logger.Debug(ctx, "Retrieved servers from registry", map[string]interface{}{
		"count": len(searchResp.Servers), "total": searchResp.Total, "query": opts.Query,
	})
	return &searchResp, nil
}

func (rc *RegistryClient) GetServer(ctx context.Context, serverID string) (*RegistryServer, error) {
	fullURL := rc.baseURL + fmt.Sprintf("/api/v1/servers/%s", url.PathEscape(serverID))
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "agent-sdk-go/1.0")
	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Failed to close response body: %v\n", closeErr)
		}
	}()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry request failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}
	var server RegistryServer
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	rc.logger.Debug(ctx, "Retrieved server from registry", map[string]interface{}{
		"server_id": serverID, "name": server.Name, "version": server.Version,
	})
	return &server, nil
}

func (rc *RegistryClient) SearchServers(ctx context.Context, query string) (*SearchResponse, error) {
	return rc.ListServers(ctx, &SearchOptions{Query: query, Limit: 50})
}

func (rc *RegistryClient) GetServersByCategory(ctx context.Context, category string) (*SearchResponse, error) {
	return rc.ListServers(ctx, &SearchOptions{Category: category, Limit: 100})
}

func (rc *RegistryClient) GetVerifiedServers(ctx context.Context) (*SearchResponse, error) {
	return rc.ListServers(ctx, &SearchOptions{Verified: true, Limit: 100})
}

func (rc *RegistryClient) GetServersByTags(ctx context.Context, tags []string) (*SearchResponse, error) {
	return rc.ListServers(ctx, &SearchOptions{Tags: tags, Limit: 100})
}
