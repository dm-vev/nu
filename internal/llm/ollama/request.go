package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// makeRequest makes an HTTP request to the Ollama API
func (c *Client) makeRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request with retry if configured
	var resp *http.Response
	if c.retryExecutor != nil {
		err = c.retryExecutor.Execute(ctx, func() error {
			var execErr error
			resp, execErr = c.HTTPClient.Do(req)
			return execErr
		})
	} else {
		resp, err = c.HTTPClient.Do(req)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
