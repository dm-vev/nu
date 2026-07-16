package vllm

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListModels lists available models
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	resp, err := c.makeGETRequest(ctx, "/v1/models")
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var modelsResponse ModelsResponse
	if err := json.Unmarshal(resp, &modelsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal models response: %w", err)
	}

	var models []string
	for _, model := range modelsResponse.Data {
		models = append(models, model.ID)
	}

	return models, nil
}
