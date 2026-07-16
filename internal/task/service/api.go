package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient is a client for making API calls.
type APIClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// APIRequest represents an API request.
type APIRequest struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
	Query   map[string]string
}

// APIResponse represents an API response.
type APIResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// NewAPIClient creates a new API client.
func NewAPIClient(baseURL string, timeout time.Duration) *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetHeader sets a header for all requests
func (c *APIClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetHeaders sets multiple headers for all requests
func (c *APIClient) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		c.headers[k] = v
	}
}

// Do makes an API request
func (c *APIClient) Do(ctx context.Context, req APIRequest) (*APIResponse, error) {
	// Prepare URL
	url := c.baseURL + req.Path

	// Prepare body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set content type if not set and body is not nil
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set query parameters
	if req.Query != nil {
		q := httpReq.URL.Query()
		for k, v := range req.Query {
			q.Add(k, v)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Make the request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer func() {
		if closeErr := httpResp.Body.Close(); closeErr != nil {
			// Log error or merge with existing error
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create response
	resp := &APIResponse{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    httpResp.Header,
	}

	return resp, nil
}

// Get makes a GET request
func (c *APIClient) Get(ctx context.Context, path string, query map[string]string, headers map[string]string) (*APIResponse, error) {
	req := APIRequest{
		Method:  http.MethodGet,
		Path:    path,
		Query:   query,
		Headers: headers,
	}
	return c.Do(ctx, req)
}

// Post makes a POST request
func (c *APIClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*APIResponse, error) {
	req := APIRequest{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	}
	return c.Do(ctx, req)
}

// Put makes a PUT request
func (c *APIClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*APIResponse, error) {
	req := APIRequest{
		Method:  http.MethodPut,
		Path:    path,
		Body:    body,
		Headers: headers,
	}
	return c.Do(ctx, req)
}

// Delete makes a DELETE request
func (c *APIClient) Delete(ctx context.Context, path string, headers map[string]string) (*APIResponse, error) {
	req := APIRequest{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	}
	return c.Do(ctx, req)
}
