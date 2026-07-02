package google

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

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Config configures a Gemini GenerateContent client.
type Config struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// Client implements Google GenerateContent streaming.
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// New constructs a Google provider client.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	return &Client{baseURL: strings.TrimRight(cfg.BaseURL, "/"), apiKey: cfg.APIKey, client: cfg.Client}
}

// BuildGenerateContentPayload builds a Gemini GenerateContent request body.
func BuildGenerateContentPayload(req provider.Request) (map[string]any, error) {
	if err := provider.ValidateRequest(req); err != nil {
		return nil, err
	}
	contents := make([]map[string]any, 0, len(req.Messages))
	for _, message := range req.Messages {
		role := message.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "tool" {
			role = "user"
		}
		contents = append(contents, map[string]any{
			"role":  role,
			"parts": []map[string]string{{"text": message.Content}},
		})
	}
	return map[string]any{"contents": contents}, nil
}

// Stream starts one Gemini streaming request.
func (c *Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error) {
	payload, err := BuildGenerateContentPayload(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.post(ctx, req.Model, payload)
	if err != nil {
		return nil, err
	}
	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		send(ctx, ch, provider.Event{Type: provider.EventStart, Provider: req.Provider, API: req.API, Model: req.Model})
		if err := parseGenerateContentStream(ctx, resp.Body, ch); err != nil {
			send(ctx, ch, provider.Event{Type: provider.EventError, ErrorClass: "fatal", Message: err.Error()})
		}
	}()
	return ch, nil
}

func (c *Client) post(ctx context.Context, model string, payload map[string]any) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("google: missing api key")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal google request: %w", err)
	}
	endpoint := c.baseURL + "/models/" + url.PathEscape(model) + ":streamGenerateContent?alt=sse&key=" + url.QueryEscape(c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create google request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send google request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("google http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

type generateContentChunk struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text         string         `json:"text"`
				FunctionCall map[string]any `json:"functionCall"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
}

func parseGenerateContentStream(ctx context.Context, body io.Reader, ch chan<- provider.Event) error {
	toolIndex := 0
	return provider.ReadSSE(ctx, body, func(ev provider.SSEvent) error {
		var chunk generateContentChunk
		if err := json.Unmarshal([]byte(ev.Data), &chunk); err != nil {
			return fmt.Errorf("decode google chunk: %w", err)
		}
		for _, candidate := range chunk.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					send(ctx, ch, provider.Event{Type: provider.EventText, Delta: part.Text})
				}
				if len(part.FunctionCall) > 0 {
					name, _ := part.FunctionCall["name"].(string)
					args, _ := json.Marshal(part.FunctionCall["args"])
					id := fmt.Sprintf("google-call-%d", toolIndex)
					send(ctx, ch, provider.Event{
						Type:       provider.EventToolCallStart,
						Index:      toolIndex,
						ToolCallID: id,
						ToolName:   name,
					})
					send(ctx, ch, provider.Event{Type: provider.EventToolCallDelta, Index: toolIndex, Delta: string(args)})
					send(ctx, ch, provider.Event{Type: provider.EventToolCallEnd, Index: toolIndex})
					toolIndex++
				}
			}
			if candidate.FinishReason != "" {
				send(ctx, ch, provider.Event{
					Type:       provider.EventDone,
					StopReason: normalizeGoogleStop(candidate.FinishReason, toolIndex > 0),
				})
			}
		}
		return nil
	})
}

func normalizeGoogleStop(reason string, usedTool bool) string {
	if usedTool {
		return "tool_use"
	}
	switch reason {
	case "MAX_TOKENS":
		return "length"
	case "SAFETY", "RECITATION":
		return "content_filter"
	default:
		return "stop"
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
