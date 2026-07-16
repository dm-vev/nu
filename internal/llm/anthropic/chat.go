package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dm-vev/nu/internal/llm"
)

// Chat uses the messages API to have a conversation with a model
func (c *Client) Chat(ctx context.Context, messages []Message, params *llm.GenerateParams) (string, error) {
	if c.Model == "" {
		return "", fmt.Errorf("model not specified: use WithModel option when creating the client")
	}
	if params == nil {
		params = llm.DefaultGenerateParams()
	}

	anthropicMessages := make([]Message, len(messages))
	var systemMessage string
	for i, msg := range messages {
		if msg.Role == "system" {
			systemMessage = msg.Content
			continue
		}
		role := msg.Role
		if role == "tool" {
			role = "assistant"
		}
		anthropicMessages[i] = Message{Role: role, Content: msg.Content}
	}
	var filteredMessages []Message
	for _, msg := range anthropicMessages {
		if msg.Role != "" && strings.TrimSpace(msg.Content) != "" {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	req := CompletionRequest{
		Model: c.Model, Messages: filteredMessages, MaxTokens: 2048,
		Temperature: params.Temperature, TopP: params.TopP, StopSequences: params.StopSequences,
	}
	if systemMessage != "" {
		req.System = systemMessage
	}
	if params.Reasoning != "" {
		c.logger.Debug(ctx, "Reasoning mode not supported in current API version", map[string]interface{}{"reasoning": params.Reasoning})
	}

	var resp CompletionResponse
	var err error
	operation := func() error {
		apiType := "Anthropic API"
		if c.VertexConfig != nil && c.VertexConfig.Enabled {
			apiType = "Vertex AI"
		}
		c.logger.Debug(ctx, "Executing "+apiType+" Chat request", map[string]interface{}{
			"model": c.Model, "temperature": req.Temperature, "top_p": req.TopP,
			"stop_sequences": req.StopSequences, "messages": len(req.Messages),
		})
		var httpReq *http.Request
		if c.VertexConfig != nil && c.VertexConfig.Enabled {
			httpReq, err = c.VertexConfig.CreateVertexHTTPRequest(ctx, &req, "POST", "/v1/messages")
			if err != nil {
				return fmt.Errorf("failed to create Vertex AI chat request: %w", err)
			}
		} else {
			reqBody, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}
			httpReq, err = http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/messages", bytes.NewBuffer(reqBody))
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("X-API-Key", c.APIKey)
			httpReq.Header.Set("Anthropic-Version", "2023-06-01")
		}

		httpResp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			c.logger.Error(ctx, "Error from Anthropic Chat API", map[string]interface{}{"error": err.Error(), "model": c.Model})
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer func() {
			if closeErr := httpResp.Body.Close(); closeErr != nil {
				c.logger.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": closeErr.Error()})
			}
		}()
		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		if httpResp.StatusCode != http.StatusOK {
			c.logger.Error(ctx, "Error from Anthropic Chat API", map[string]interface{}{
				"status_code": httpResp.StatusCode, "response": string(respBody), "model": c.Model,
			})
			return fmt.Errorf("error from Anthropic API: %s", string(respBody))
		}
		c.logger.Debug(ctx, "Raw streaming response before unmarshaling", map[string]interface{}{
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
		})
		if err = json.Unmarshal(respBody, &resp); err != nil {
			c.logger.Error(ctx, "Failed to unmarshal streaming response", map[string]interface{}{
				"error": err.Error(), "response_length": len(respBody),
				"response_sample": func() string {
					if len(respBody) > 200 {
						return string(respBody[:200])
					}
					return string(respBody)
				}(),
			})
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return nil
	}

	if c.vertexRetryExecutor != nil {
		c.logger.Info(ctx, "Using Vertex retry mechanism with region rotation for Chat", map[string]interface{}{
			"model": c.Model, "current_region": c.VertexConfig.GetCurrentRegion(),
		})
		err = c.vertexRetryExecutor.Execute(ctx, operation)
	} else if c.retryExecutor != nil {
		c.logger.Info(ctx, "Using standard retry mechanism for Anthropic Chat request", map[string]interface{}{
			"model": c.Model, "vertex_config_available": c.VertexConfig != nil,
		})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		c.logger.Debug(ctx, "No retry mechanism configured for Chat request", map[string]interface{}{"model": c.Model})
		err = operation()
	}
	if err != nil {
		return "", err
	}

	var contentText []string
	for _, block := range resp.Content {
		if block.Type == "text" {
			contentText = append(contentText, block.Text)
		}
	}
	if len(contentText) == 0 {
		return "", fmt.Errorf("no text content in response")
	}
	response := strings.Join(contentText, "\n")
	c.logger.Debug(ctx, "Successfully received chat response from Anthropic", map[string]interface{}{
		"model": c.Model, "response_length": len(response), "response_preview": response,
	})
	return response, nil
}
