package anthropic

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dm-vev/nu/contracts"
)

func (c *Client) executeStreamingRequestWithMemory(ctx context.Context, req CompletionRequest, eventChan chan<- contracts.StreamEvent, prompt string, params *contracts.GenerateOptions) error {
	operation := func() error {
		c.logger.Debug(ctx, "Executing Anthropic streaming API request", map[string]interface{}{
			"model": c.Model, "temperature": req.Temperature, "top_p": req.TopP,
			"stop_sequences": req.StopSequences, "system": req.System != "", "stream": req.Stream,
		})
		if c.BedrockConfig != nil && c.BedrockConfig.Enabled {
			var cacheConfig *contracts.CacheConfig
			if params != nil {
				cacheConfig = params.CacheConfig
			}
			return c.executeBedrockStreaming(ctx, &req, eventChan, cacheConfig)
		}

		httpReq, err := c.createStreamingHTTPRequest(ctx, &req, "/v1/messages")
		if err != nil {
			return fmt.Errorf("failed to create streaming request: %w", err)
		}
		httpResp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			c.logger.Error(ctx, "Error from Anthropic streaming API", map[string]interface{}{"error": err.Error(), "model": c.Model})
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer func() {
			if closeErr := httpResp.Body.Close(); closeErr != nil {
				c.logger.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": closeErr.Error()})
			}
		}()

		if httpResp.StatusCode != http.StatusOK {
			var errorBody []byte
			if httpResp.Body != nil {
				errorBody, _ = io.ReadAll(httpResp.Body)
				_ = httpResp.Body.Close()
			}
			c.logger.Error(ctx, "Error from Anthropic streaming API", map[string]interface{}{
				"status_code": httpResp.StatusCode, "model": c.Model,
				"error_body": string(errorBody), "content_type": httpResp.Header.Get("Content-Type"),
			})
			if len(errorBody) > 0 {
				return fmt.Errorf("error from Anthropic API: HTTP %d - %s", httpResp.StatusCode, string(errorBody))
			}
			return fmt.Errorf("error from Anthropic API: HTTP %d", httpResp.StatusCode)
		}

		contentType := httpResp.Header.Get("Content-Type")
		if contentType != "text/event-stream" && contentType != "text/event-stream; charset=utf-8" {
			return fmt.Errorf("unexpected content type: %s", contentType)
		}

		scanner := bufio.NewScanner(httpResp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		_ = c.parseSSEStreamAndCapture(ctx, scanner, eventChan, req, prompt, params)
		if err := scanner.Err(); err != nil {
			c.logger.Error(ctx, "Scanner error while reading SSE stream", map[string]interface{}{"error": err.Error(), "model": c.Model})
			return fmt.Errorf("scanner error while reading SSE stream: %w", err)
		}
		c.logger.Debug(ctx, "Successfully completed Anthropic streaming request", map[string]interface{}{"model": c.Model})
		return nil
	}

	if c.vertexRetryExecutor != nil {
		c.logger.Info(ctx, "Using Vertex retry mechanism with region rotation for streaming", map[string]interface{}{
			"model": c.Model, "current_region": c.VertexConfig.GetCurrentRegion(),
		})
		return c.vertexRetryExecutor.Execute(ctx, operation)
	} else if c.retryExecutor != nil {
		c.logger.Info(ctx, "Using standard retry mechanism for Anthropic streaming request", map[string]interface{}{
			"model": c.Model, "vertex_config_available": c.VertexConfig != nil,
		})
		return c.retryExecutor.Execute(ctx, operation)
	}
	c.logger.Debug(ctx, "No retry mechanism configured for streaming", map[string]interface{}{"model": c.Model})
	return operation()
}
