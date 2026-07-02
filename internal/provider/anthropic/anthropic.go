package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"nu/internal/provider"
)

const (
	defaultBaseURL          = "https://api.anthropic.com"
	defaultAnthropicVersion = "2023-06-01"
	defaultMaxTokens        = 4096
)

// Config configures an Anthropic Messages client.
type Config struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// Client implements Anthropic Messages streaming.
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// New constructs an Anthropic client.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		client:  cfg.Client,
	}
}

// BuildMessagesPayload builds an Anthropic Messages request body.
func BuildMessagesPayload(req provider.Request) (map[string]any, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	messages := make([]map[string]any, 0, len(req.Messages))
	for _, message := range req.Messages {
		switch message.Role {
		case "tool":
			messages = append(messages, map[string]any{
				"role": "user",
				"content": []map[string]any{{
					"type":        "tool_result",
					"tool_use_id": message.ToolCallID,
					"content":     message.Content,
				}},
			})
		case "assistant", "user":
			messages = append(messages, map[string]any{"role": message.Role, "content": message.Content})
		default:
			messages = append(messages, map[string]any{"role": "user", "content": message.Content})
		}
	}
	return map[string]any{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": defaultMaxTokens,
		"stream":     true,
	}, nil
}

// Stream starts one Anthropic streaming request.
func (c *Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	payload, err := BuildMessagesPayload(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.post(ctx, payload)
	if err != nil {
		return nil, err
	}
	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		send(ctx, ch, provider.Event{Type: provider.EventStart, Provider: req.Provider, API: req.API, Model: req.Model})
		if err := parseMessagesStream(ctx, resp.Body, ch); err != nil {
			send(ctx, ch, provider.Event{Type: provider.EventError, ErrorClass: "fatal", Message: err.Error()})
		}
	}()
	return ch, nil
}

func (c *Client) post(ctx context.Context, payload map[string]any) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("anthropic: missing api key")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}
	endpoint, err := url.JoinPath(c.baseURL, "v1/messages")
	if err != nil {
		return nil, fmt.Errorf("build anthropic endpoint: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create anthropic request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", defaultAnthropicVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send anthropic request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("anthropic http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

type contentBlockStart struct {
	Index   int `json:"index"`
	Content struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"content_block"`
}

type contentBlockDelta struct {
	Index int `json:"index"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
		Thinking    string `json:"thinking"`
	} `json:"delta"`
}

type messageDelta struct {
	Delta struct {
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
}

type streamError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func parseMessagesStream(ctx context.Context, body io.Reader, ch chan<- provider.Event) error {
	started := map[int]bool{}
	stopReason := "stop"
	return provider.ReadSSE(ctx, body, func(ev provider.SSEvent) error {
		switch ev.Event {
		case "content_block_start":
			var start contentBlockStart
			if err := json.Unmarshal([]byte(ev.Data), &start); err != nil {
				return fmt.Errorf("decode anthropic block start: %w", err)
			}
			if start.Content.Type == "tool_use" {
				started[start.Index] = true
				send(ctx, ch, provider.Event{
					Type:       provider.EventToolCallStart,
					Index:      start.Index,
					ToolCallID: start.Content.ID,
					ToolName:   start.Content.Name,
				})
			}
		case "content_block_delta":
			var delta contentBlockDelta
			if err := json.Unmarshal([]byte(ev.Data), &delta); err != nil {
				return fmt.Errorf("decode anthropic block delta: %w", err)
			}
			switch delta.Delta.Type {
			case "text_delta":
				send(ctx, ch, provider.Event{Type: provider.EventText, Index: delta.Index, Delta: delta.Delta.Text})
			case "input_json_delta":
				send(ctx, ch, provider.Event{
					Type:  provider.EventToolCallDelta,
					Index: delta.Index,
					Delta: delta.Delta.PartialJSON,
				})
			}
		case "content_block_stop":
			var stop struct {
				Index int `json:"index"`
			}
			if err := json.Unmarshal([]byte(ev.Data), &stop); err != nil {
				return fmt.Errorf("decode anthropic block stop: %w", err)
			}
			if started[stop.Index] {
				send(ctx, ch, provider.Event{Type: provider.EventToolCallEnd, Index: stop.Index})
			}
		case "message_delta":
			var delta messageDelta
			if err := json.Unmarshal([]byte(ev.Data), &delta); err != nil {
				return fmt.Errorf("decode anthropic message delta: %w", err)
			}
			if delta.Delta.StopReason != "" {
				stopReason = normalizeAnthropicStop(delta.Delta.StopReason)
			}
		case "message_stop":
			send(ctx, ch, provider.Event{Type: provider.EventDone, StopReason: stopReason})
		case "error":
			var value streamError
			if err := json.Unmarshal([]byte(ev.Data), &value); err != nil {
				return fmt.Errorf("decode anthropic stream error: %w", err)
			}
			send(ctx, ch, provider.Event{
				Type:       provider.EventError,
				ErrorClass: normalizeAnthropicError(value.Error.Type),
				Message:    value.Error.Message,
			})
		}
		return nil
	})
}

func normalizeAnthropicStop(reason string) string {
	switch reason {
	case "tool_use":
		return "tool_use"
	case "max_tokens":
		return "length"
	default:
		return "stop"
	}
}

func normalizeAnthropicError(kind string) string {
	switch kind {
	case "authentication_error", "permission_error":
		return "auth"
	case "rate_limit_error", "overloaded_error":
		return "rate_limit"
	default:
		return "fatal"
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
