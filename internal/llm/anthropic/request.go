package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"nu/internal/contracts"
)

func (c *Client) createHTTPRequest(ctx context.Context, req *CompletionRequest, path string) (*http.Request, error) {
	return c.createHTTPRequestWithCache(ctx, req, path, nil)
}

func (c *Client) createHTTPRequestWithCache(ctx context.Context, req *CompletionRequest, path string, cacheConfig *contracts.CacheConfig) (*http.Request, error) {
	if c.VertexConfig != nil && c.VertexConfig.Enabled {
		return c.VertexConfig.CreateVertexHTTPRequest(ctx, req, "POST", path)
	}

	var reqBody []byte
	cacheBuilder := anthropicNewCacheRequestBuilder(cacheConfig)
	if cacheBuilder.HasCacheOptions() {
		cacheableReq, err := cacheBuilder.BuildCacheableRequest(req)
		if err != nil {
			return nil, fmt.Errorf("failed to build cacheable request: %w", err)
		}
		reqBody, err = json.Marshal(cacheableReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cacheable request: %w", err)
		}
	} else {
		var err error
		reqBody, err = json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.APIKey)
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")
	return httpReq, nil
}

func (c *Client) createStreamingHTTPRequest(ctx context.Context, req *CompletionRequest, path string) (*http.Request, error) {
	return c.createStreamingHTTPRequestWithCache(ctx, req, path, nil)
}

func (c *Client) createStreamingHTTPRequestWithCache(ctx context.Context, req *CompletionRequest, path string, cacheConfig *contracts.CacheConfig) (*http.Request, error) {
	if c.VertexConfig != nil && c.VertexConfig.Enabled {
		return c.VertexConfig.CreateVertexStreamingHTTPRequest(ctx, req, "POST", path)
	}

	req.Stream = true
	var reqBody []byte
	cacheBuilder := anthropicNewCacheRequestBuilder(cacheConfig)
	if cacheBuilder.HasCacheOptions() {
		cacheableReq, err := cacheBuilder.BuildCacheableRequest(req)
		if err != nil {
			return nil, fmt.Errorf("failed to build cacheable streaming request: %w", err)
		}
		cacheableReq.Stream = true
		reqBody, err = json.Marshal(cacheableReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cacheable streaming request: %w", err)
		}
	} else {
		var err error
		reqBody, err = json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.APIKey)
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	return httpReq, nil
}
