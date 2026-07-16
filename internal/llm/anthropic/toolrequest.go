package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dm-vev/nu/contracts"
)

func (c *Client) invokeToolRequest(ctx context.Context, req *CompletionRequest, params *contracts.GenerateOptions, iteration int) (*CompletionResponse, error) {
	var resp CompletionResponse
	var err error
	operation := func() error {
		if c.BedrockConfig != nil && c.BedrockConfig.Enabled {
			bedrockResp, err := c.BedrockConfig.InvokeModel(ctx, c.Model, req, params.CacheConfig)
			if err != nil {
				return fmt.Errorf("failed to invoke Bedrock model (iteration %d): %w", iteration+1, err)
			}
			resp = *bedrockResp
			return nil
		}

		httpReq, err := c.createHTTPRequestWithCache(ctx, req, "/v1/messages", params.CacheConfig)
		if err != nil {
			return fmt.Errorf("failed to create request (iteration %d): %w", iteration+1, err)
		}
		httpResp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			c.logger.Error(ctx, "Error from Anthropic API", map[string]interface{}{
				"error": err.Error(), "model": c.Model, "iteration": iteration + 1,
			})
			return fmt.Errorf("failed to send request (iteration %d): %w", iteration+1, err)
		}
		defer func() {
			if closeErr := httpResp.Body.Close(); closeErr != nil {
				c.logger.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": closeErr.Error()})
			}
		}()
		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body (iteration %d): %w", iteration+1, err)
		}
		if httpResp.StatusCode != http.StatusOK {
			c.logger.Error(ctx, "Error from Anthropic API", map[string]interface{}{
				"status_code": httpResp.StatusCode, "response": string(respBody),
				"model": c.Model, "iteration": iteration + 1,
			})
			return fmt.Errorf("error from Anthropic API (iteration %d): %s", iteration+1, string(respBody))
		}
		c.logger.Debug(ctx, "Raw response before unmarshaling", map[string]interface{}{
			"response_length": len(respBody),
			"response_prefix": func() string {
				if len(respBody) > 100 {
					return string(respBody[:100])
				}
				return string(respBody)
			}(),
			"first_char": func() string {
				if len(respBody) > 0 {
					return fmt.Sprintf("'%c' (0x%02x)", respBody[0], respBody[0])
				}
				return "empty"
			}(),
			"iteration": iteration + 1,
		})
		if err = json.Unmarshal(respBody, &resp); err != nil {
			return fmt.Errorf("failed to unmarshal response (iteration %d): %w", iteration+1, err)
		}
		return nil
	}

	if c.vertexRetryExecutor != nil {
		c.logger.Info(ctx, "Using Vertex retry mechanism with region rotation for GenerateWithTools", map[string]interface{}{
			"model": c.Model, "current_region": c.VertexConfig.GetCurrentRegion(), "iteration": iteration + 1,
		})
		err = c.vertexRetryExecutor.Execute(ctx, operation)
	} else if c.retryExecutor != nil {
		c.logger.Info(ctx, "Using standard retry mechanism for GenerateWithTools", map[string]interface{}{
			"model": c.Model, "vertex_config_available": c.VertexConfig != nil, "iteration": iteration + 1,
		})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		c.logger.Debug(ctx, "No retry mechanism configured for GenerateWithTools", map[string]interface{}{
			"model": c.Model, "iteration": iteration + 1,
		})
		err = operation()
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) inspectToolResponse(ctx context.Context, resp *CompletionResponse, iteration int) (bool, []ToolUse, []string) {
	var hasToolUse bool
	var toolCalls []ToolUse
	var textContent []string
	c.logger.Debug(ctx, "Response content blocks", map[string]interface{}{
		"numBlocks": len(resp.Content), "iteration": iteration + 1,
		"blockTypes": func() []string {
			types := make([]string, len(resp.Content))
			for i, block := range resp.Content {
				types[i] = block.Type
				if block.Type == "tool_use" && block.ToolUse != nil {
					toolName := block.ToolUse.Name
					if toolName == "" {
						toolName = block.ToolUse.RecipientName
					}
					c.logger.Debug(ctx, "Found tool use block", map[string]interface{}{
						"toolName": toolName, "toolID": block.ToolUse.ID, "iteration": iteration + 1,
					})
				}
			}
			return types
		}(),
	})

	for _, contentBlock := range resp.Content {
		switch contentBlock.Type {
		case "tool_use":
			hasToolUse = true
			if contentBlock.ToolUse != nil {
				toolCalls = append(toolCalls, *contentBlock.ToolUse)
			} else if contentBlock.ID != "" && contentBlock.Name != "" {
				toolCalls = append(toolCalls, ToolUse{ID: contentBlock.ID, Name: contentBlock.Name, Input: contentBlock.Input})
			}
		case "text":
			textContent = append(textContent, contentBlock.Text)
		}
	}
	c.logger.Debug(ctx, "Tool use detection results", map[string]interface{}{
		"hasToolUse": hasToolUse, "toolCalls": len(toolCalls), "iteration": iteration + 1,
	})
	return hasToolUse, toolCalls, textContent
}
