package ollama

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListModels lists available models
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	resp, err := c.makeRequest(ctx, "/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var tagsResponse struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(resp, &tagsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal models response: %w", err)
	}

	var models []string
	for _, model := range tagsResponse.Models {
		models = append(models, model.Name)
	}

	return models, nil
}

// PullModel downloads a model
func (c *Client) PullModel(ctx context.Context, modelName string) error {
	req := struct {
		Name string `json:"name"`
	}{
		Name: modelName,
	}

	_, err := c.makeRequest(ctx, "/api/pull", req)
	if err != nil {
		return fmt.Errorf("failed to pull model %s: %w", modelName, err)
	}

	return nil
}
