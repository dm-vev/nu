package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ChatCompletionRequest represents a request to the DeepSeek Chat Completion API
type ChatCompletionRequest struct {
	Model            string               `json:"model"`
	Messages         []Message            `json:"messages"`
	Temperature      float64              `json:"temperature,omitempty"`
	TopP             float64              `json:"top_p,omitempty"`
	FrequencyPenalty float64              `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64              `json:"presence_penalty,omitempty"`
	Stop             []string             `json:"stop,omitempty"`
	MaxTokens        int                  `json:"max_tokens,omitempty"`
	Stream           bool                 `json:"stream,omitempty"`
	Tools            []Tool               `json:"tools,omitempty"`
	ToolChoice       interface{}          `json:"tool_choice,omitempty"`
	ResponseFormat   *ResponseFormatParam `json:"response_format,omitempty"`
}

// DeepSeekMessage represents a message in the chat
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall represents a tool call in the response
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// DeepSeekTool represents a tool/function definition
type Tool struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// FunctionDef represents a function definition
type FunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ResponseFormatParam represents the response format parameter
type ResponseFormatParam struct {
	Type       string      `json:"type"`
	JSONSchema interface{} `json:"json_schema,omitempty"`
}

// ChatCompletionResponse represents a response from the DeepSeek Chat Completion API
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// DeepSeekUsage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// doRequest performs an HTTP request to the DeepSeek API
func (c *Client) doRequest(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/chat/completions", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Make request
	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			c.logger.Error(ctx, "Failed to close response body", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DeepSeek API error: status=%d, body=%s", httpResp.StatusCode, string(body))
	}

	// Parse response
	var resp ChatCompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}
