package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// TransformRequest converts an Anthropic request to Vertex AI format
// Returns the full URL, headers, and modified request body
func (vc *VertexConfig) TransformRequest(req *CompletionRequest, method, path string) (string, map[string]string, []byte, error) {
	if !vc.Enabled {
		return "", nil, nil, fmt.Errorf("vertex AI is not enabled")
	}

	model := req.Model
	if model == "" {
		return "", nil, nil, fmt.Errorf("model is required for Vertex AI")
	}

	vertexReq := *req
	vertexReq.Model = ""
	vertexReq.Version = "vertex-2023-10-16"

	var endpoint string
	if strings.Contains(path, "messages") {
		if req.Stream {
			endpoint = "streamRawPredict"
		} else {
			endpoint = "rawPredict"
		}
	} else {
		endpoint = "rawPredict"
	}

	url := fmt.Sprintf(
		"%s/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:%s",
		vc.GetBaseURL(),
		vc.ProjectID,
		vc.GetCurrentRegion(),
		model,
		endpoint,
	)

	headers := map[string]string{"Content-Type": "application/json"}
	reqBody, err := json.Marshal(vertexReq)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to marshal Vertex AI request: %w", err)
	}

	return url, headers, reqBody, nil
}

// CreateVertexHTTPRequest creates an HTTP request configured for Vertex AI
func (vc *VertexConfig) CreateVertexHTTPRequest(ctx context.Context, req *CompletionRequest, method, path string) (*http.Request, error) {
	if !vc.Enabled {
		return nil, fmt.Errorf("vertex AI is not enabled")
	}

	url, headers, body, err := vc.TransformRequest(req, method, path)
	if err != nil {
		return nil, fmt.Errorf("failed to transform request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	authHeaders, err := vc.GetAuthHeaders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth headers: %w", err)
	}
	for key, value := range authHeaders {
		httpReq.Header.Set(key, value)
	}

	if req.Stream {
		httpReq.Header.Set("Accept", "text/event-stream")
		httpReq.Header.Set("Cache-Control", "no-cache")
	}

	return httpReq, nil
}

// CreateVertexStreamingHTTPRequest creates an HTTP request configured for Vertex AI streaming
func (vc *VertexConfig) CreateVertexStreamingHTTPRequest(ctx context.Context, req *CompletionRequest, method, path string) (*http.Request, error) {
	req.Stream = true
	httpReq, err := vc.CreateVertexHTTPRequest(ctx, req, method, path)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	return httpReq, nil
}
