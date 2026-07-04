package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"nu/internal/provider"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Config configures an OpenAI provider client.
type Config struct {
	BaseURL string
	APIKey  string
	API     string
	Client  *http.Client
}

// Client implements OpenAI Chat Completions and Responses streaming.
type Client struct {
	baseURL string
	apiKey  string
	api     string
	client  *http.Client
}

// New constructs an OpenAI client.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.API == "" {
		cfg.API = "responses"
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		api:     cfg.API,
		client:  cfg.Client,
	}
}

// BuildChatPayload builds a Chat Completions request body.
func BuildChatPayload(req provider.Request) (map[string]any, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	messages := make([]map[string]any, 0, len(req.Messages))
	for _, message := range req.Messages {
		messages = append(messages, chatMessage(message))
	}
	payload := map[string]any{
		"model":          req.Model,
		"messages":       messages,
		"stream":         true,
		"stream_options": map[string]any{"include_usage": true},
	}
	if len(req.Tools) > 0 {
		payload["tools"] = chatTools(req.Tools)
		payload["tool_choice"] = "auto"
	}
	return payload, nil
}

func chatMessage(message provider.Message) map[string]any {
	if message.Role == "assistant" && message.ToolCallID != "" {
		return map[string]any{
			"role":    "assistant",
			"content": nil,
			"tool_calls": []map[string]any{{
				"id":   message.ToolCallID,
				"type": "function",
				"function": map[string]string{
					"name":      message.Name,
					"arguments": message.Content,
				},
			}},
		}
	}
	item := map[string]any{"role": message.Role, "content": message.Content}
	if message.Role == "tool" {
		item["tool_call_id"] = message.ToolCallID
	}
	if message.Name != "" {
		item["name"] = message.Name
	}
	return item
}

// BuildResponsesPayload builds a Responses request body.
func BuildResponsesPayload(req provider.Request) (map[string]any, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	input := make([]map[string]any, 0, len(req.Messages))
	for _, message := range req.Messages {
		input = append(input, responsesInput(message))
	}
	payload := map[string]any{
		"model":  req.Model,
		"input":  input,
		"stream": true,
	}
	if len(req.Tools) > 0 {
		payload["tools"] = responsesTools(req.Tools)
	}
	return payload, nil
}

func chatTools(tools []provider.ToolDefinition) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}
	return out
}

func responsesTools(tools []provider.ToolDefinition) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		out = append(out, map[string]any{
			"type":        "function",
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		})
	}
	return out
}

func responsesInput(message provider.Message) map[string]any {
	if message.Role == "assistant" && message.ToolCallID != "" {
		return map[string]any{
			"type":      "function_call",
			"call_id":   message.ToolCallID,
			"name":      message.Name,
			"arguments": message.Content,
		}
	}
	if message.Role == "tool" {
		return map[string]any{
			"type":    "function_call_output",
			"call_id": message.ToolCallID,
			"output":  message.Content,
		}
	}
	return map[string]any{"role": message.Role, "content": message.Content}
}

// Stream starts one OpenAI streaming request.
func (c *Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	api := c.api
	if req.API != "" {
		api = req.API
	}
	payload, path, err := c.payloadAndPath(req, api)
	if err != nil {
		return nil, err
	}
	resp, err := c.post(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		send(ctx, ch, provider.Event{Type: provider.EventStart, Provider: req.Provider, API: req.API, Model: req.Model})
		var parseErr error
		if api == "chat" {
			parseErr = parseChatStream(ctx, resp.Body, ch)
		} else {
			parseErr = parseResponsesStream(ctx, resp.Body, ch)
		}
		if parseErr != nil {
			send(ctx, ch, provider.Event{Type: provider.EventError, ErrorClass: "fatal", Message: parseErr.Error()})
		}
	}()
	return ch, nil
}

func (c *Client) payloadAndPath(req provider.Request, api string) (map[string]any, string, error) {
	switch api {
	case "chat":
		payload, err := BuildChatPayload(req)
		return payload, "chat/completions", err
	case "responses":
		payload, err := BuildResponsesPayload(req)
		return payload, "responses", err
	default:
		return nil, "", fmt.Errorf("openai api %q is not supported", api)
	}
}

func (c *Client) post(ctx context.Context, path string, payload map[string]any) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("openai: missing api key")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}
	endpoint, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("build openai endpoint: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send openai request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("%w: openai http %d: %s", provider.ErrRateLimit, resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return nil, fmt.Errorf("openai http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

type chatChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func parseChatStream(ctx context.Context, body io.Reader, ch chan<- provider.Event) error {
	started := map[int]bool{}
	return provider.ReadSSE(ctx, body, func(ev provider.SSEvent) error {
		if strings.TrimSpace(ev.Data) == "[DONE]" {
			return nil
		}
		var chunk chatChunk
		if err := json.Unmarshal([]byte(ev.Data), &chunk); err != nil {
			return fmt.Errorf("decode openai chat chunk: %w", err)
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				send(ctx, ch, provider.Event{Type: provider.EventText, Delta: choice.Delta.Content})
			}
			for _, toolCall := range choice.Delta.ToolCalls {
				if !started[toolCall.Index] {
					if toolCall.ID == "" || toolCall.Function.Name == "" {
						return fmt.Errorf("openai chat tool call missing id or name at index %d", toolCall.Index)
					}
					started[toolCall.Index] = true
					send(ctx, ch, provider.Event{
						Type:       provider.EventToolCallStart,
						Index:      toolCall.Index,
						ToolCallID: toolCall.ID,
						ToolName:   toolCall.Function.Name,
					})
				}
				if toolCall.Function.Arguments != "" {
					send(ctx, ch, provider.Event{
						Type:  provider.EventToolCallDelta,
						Index: toolCall.Index,
						Delta: toolCall.Function.Arguments,
					})
				}
			}
			if choice.FinishReason != nil {
				stopReason := normalizeOpenAIStop(*choice.FinishReason)
				sendToolEnds(ctx, ch, started)
				send(ctx, ch, provider.Event{Type: provider.EventDone, StopReason: stopReason})
			}
		}
		return nil
	})
}

type responseItemEvent struct {
	OutputIndex int `json:"output_index"`
	Item        struct {
		Type   string `json:"type"`
		ID     string `json:"id"`
		CallID string `json:"call_id"`
		Name   string `json:"name"`
	} `json:"item"`
}

type responseDeltaEvent struct {
	OutputIndex int    `json:"output_index"`
	Delta       string `json:"delta"`
}

func parseResponsesStream(ctx context.Context, body io.Reader, ch chan<- provider.Event) error {
	started := map[int]bool{}
	sawTool := false
	return provider.ReadSSE(ctx, body, func(ev provider.SSEvent) error {
		switch ev.Event {
		case "response.output_text.delta":
			var delta responseDeltaEvent
			if err := json.Unmarshal([]byte(ev.Data), &delta); err != nil {
				return fmt.Errorf("decode openai response text delta: %w", err)
			}
			if delta.Delta != "" {
				send(ctx, ch, provider.Event{Type: provider.EventText, Index: delta.OutputIndex, Delta: delta.Delta})
			}
		case "response.output_item.added":
			var item responseItemEvent
			if err := json.Unmarshal([]byte(ev.Data), &item); err != nil {
				return fmt.Errorf("decode openai response item: %w", err)
			}
			if item.Item.Type == "function_call" {
				sawTool = true
				started[item.OutputIndex] = true
				send(ctx, ch, provider.Event{
					Type:       provider.EventToolCallStart,
					Index:      item.OutputIndex,
					ToolCallID: firstNonEmpty(item.Item.CallID, item.Item.ID),
					ToolName:   item.Item.Name,
				})
			}
		case "response.function_call_arguments.delta":
			var delta responseDeltaEvent
			if err := json.Unmarshal([]byte(ev.Data), &delta); err != nil {
				return fmt.Errorf("decode openai response tool delta: %w", err)
			}
			send(ctx, ch, provider.Event{Type: provider.EventToolCallDelta, Index: delta.OutputIndex, Delta: delta.Delta})
		case "response.output_item.done":
			var item responseItemEvent
			if err := json.Unmarshal([]byte(ev.Data), &item); err != nil {
				return fmt.Errorf("decode openai response item done: %w", err)
			}
			if started[item.OutputIndex] {
				send(ctx, ch, provider.Event{Type: provider.EventToolCallEnd, Index: item.OutputIndex})
			}
		case "response.completed":
			stopReason := "stop"
			if sawTool {
				stopReason = "tool_use"
			}
			send(ctx, ch, provider.Event{Type: provider.EventDone, StopReason: stopReason})
		case "response.failed", "error":
			send(ctx, ch, provider.Event{Type: provider.EventError, ErrorClass: normalizeOpenAIError(ev.Data), Message: ev.Data})
		}
		return nil
	})
}

func normalizeOpenAIError(value string) string {
	if strings.Contains(strings.ToLower(value), "rate") {
		return "rate_limit"
	}
	return "fatal"
}

func normalizeOpenAIStop(reason string) string {
	switch reason {
	case "tool_calls":
		return "tool_use"
	case "length":
		return "length"
	case "content_filter":
		return "content_filter"
	default:
		return "stop"
	}
}

func sendToolEnds(ctx context.Context, ch chan<- provider.Event, started map[int]bool) {
	indexes := make([]int, 0, len(started))
	for index := range started {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		send(ctx, ch, provider.Event{Type: provider.EventToolCallEnd, Index: index})
	}
}

func send(ctx context.Context, ch chan<- provider.Event, ev provider.Event) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- ev:
		return true
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
