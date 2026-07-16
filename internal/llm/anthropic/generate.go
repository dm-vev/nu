package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// Generate generates text from a prompt
func (c *Client) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	response, err := c.generateInternal(ctx, prompt, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

func (c *Client) generateInternal(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	if c.Model == "" {
		return nil, fmt.Errorf("model not specified: use WithModel option when creating the client")
	}
	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, option := range options {
		option(params)
	}
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, "default")
	}
	messages := c.buildMessagesWithMemory(ctx, prompt, params)
	if params.ResponseFormat != nil {
		schemaJSON, err := json.MarshalIndent(params.ResponseFormat.Schema, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response format schema: %w", err)
		}
		exampleJSON := anthropicCreateExampleFromSchema(params.ResponseFormat.Schema)
		exampleStr, _ := json.MarshalIndent(exampleJSON, "", "  ")
		messages[0].Content = fmt.Sprintf(`%s

You must respond with a valid JSON object that exactly follows this schema:
%s

Example output:
%s

Return only the JSON object, with no additional text or markdown formatting.`, prompt, string(schemaJSON), string(exampleStr))
	}

	maxTokens := 2048
	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && params.LLMConfig.ReasoningBudget > 0 {
		maxTokens = params.LLMConfig.ReasoningBudget + 4000
	}
	req := CompletionRequest{
		Model: c.Model, Messages: messages, MaxTokens: maxTokens,
		Temperature: params.LLMConfig.Temperature, TopP: params.LLMConfig.TopP,
	}
	if len(params.LLMConfig.StopSequences) > 0 {
		req.StopSequences = params.LLMConfig.StopSequences
	}
	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && SupportsThinking(c.Model) {
		req.Thinking = &ReasoningSpec{Type: "enabled"}
		if params.LLMConfig.ReasoningBudget > 0 {
			req.Thinking.BudgetTokens = params.LLMConfig.ReasoningBudget
		}
		req.Temperature = 1.0
		c.logger.Debug(ctx, "Enabled reasoning (thinking) tokens", map[string]interface{}{
			"model": c.Model, "budget_tokens": params.LLMConfig.ReasoningBudget,
			"max_tokens": req.MaxTokens, "temperature": req.Temperature,
		})
	} else if params.LLMConfig != nil && params.LLMConfig.EnableReasoning {
		c.logger.Warn(ctx, "Thinking tokens not supported by this model", map[string]interface{}{
			"model":            c.Model,
			"supported_models": []string{"claude-3-7-sonnet-20250219", "claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-opus-4-1-20250805"},
		})
	}
	if params.SystemMessage != "" {
		req.System = params.SystemMessage
	}

	var resp CompletionResponse
	var err error
	operation := func() error {
		var apiType string
		if c.BedrockConfig != nil && c.BedrockConfig.Enabled {
			apiType = "bedrock"
		} else if c.VertexConfig != nil && c.VertexConfig.Enabled {
			apiType = "vertex"
		} else {
			apiType = "anthropic"
		}
		if c.BedrockConfig != nil && c.BedrockConfig.Enabled {
			bedrockResp, err := c.BedrockConfig.InvokeModel(ctx, c.Model, &req, params.CacheConfig)
			if err != nil {
				return fmt.Errorf("failed to invoke Bedrock model: %w", err)
			}
			resp = *bedrockResp
			return nil
		}

		var httpReq *http.Request
		if c.VertexConfig != nil && c.VertexConfig.Enabled {
			httpReq, err = c.VertexConfig.CreateVertexHTTPRequest(ctx, &req, "POST", "/v1/messages")
			if err != nil {
				return fmt.Errorf("failed to create Vertex AI request: %w", err)
			}
		} else {
			var reqBody []byte
			cacheBuilder := anthropicNewCacheRequestBuilder(params.CacheConfig)
			if cacheBuilder.HasCacheOptions() {
				cacheableReq, err := cacheBuilder.BuildCacheableRequest(&req)
				if err != nil {
					return fmt.Errorf("failed to build cacheable request: %w", err)
				}
				reqBody, err = json.Marshal(cacheableReq)
				if err != nil {
					return fmt.Errorf("failed to marshal cacheable request: %w", err)
				}
			} else {
				var err error
				reqBody, err = json.Marshal(req)
				if err != nil {
					return fmt.Errorf("failed to marshal request: %w", err)
				}
			}
			httpReq, err = http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/messages", bytes.NewBuffer(reqBody))
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("X-API-Key", c.APIKey)
			httpReq.Header.Set("anthropic-version", "2023-06-01")
		}

		httpResp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("failed to send request to %s: %w", apiType, err)
		}
		defer func() {
			if err := httpResp.Body.Close(); err != nil {
				fmt.Printf("Warning: failed to close response body: %v\n", err)
			}
		}()
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read %s response: %w", apiType, err)
		}
		if httpResp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s API error (status %d): %s", apiType, httpResp.StatusCode, string(body))
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse %s response: %w", apiType, err)
		}
		return nil
	}

	if c.VertexConfig != nil && c.VertexConfig.Enabled && c.vertexRetryExecutor != nil {
		err = c.vertexRetryExecutor.Execute(ctx, operation)
	} else if c.retryExecutor != nil {
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		err = operation()
	}
	if err != nil {
		return nil, err
	}

	var contentText []string
	for _, block := range resp.Content {
		if block.Type == "text" {
			contentText = append(contentText, block.Text)
		}
	}
	if len(contentText) == 0 {
		return nil, fmt.Errorf("no text content in response")
	}
	content := strings.Join(contentText, "\n")
	if params.ResponseFormat != nil && !strings.HasPrefix(strings.TrimSpace(content), "{") {
		content = "{" + content
	}
	return &contracts.LLMResponse{
		Content: content, Model: resp.Model, StopReason: resp.StopReason,
		Usage: &contracts.TokenUsage{
			InputTokens: resp.Usage.InputTokens, OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:              resp.Usage.InputTokens + resp.Usage.OutputTokens,
			CacheCreationInputTokens: resp.Usage.CacheCreationInputTokens,
			CacheReadInputTokens:     resp.Usage.CacheReadInputTokens,
		},
		Metadata: map[string]interface{}{"provider": "anthropic"},
	}, nil
}

// GenerateDetailed generates text and returns detailed response information including token usage
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	return c.generateInternal(ctx, prompt, options...)
}
